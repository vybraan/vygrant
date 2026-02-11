package certgen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GenerateSelfSignedCert() (tls.Certificate, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return tls.Certificate{}, "", err
	}

	certDir := filepath.Join(home, ".vybr", "vygrant", "certs")
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return tls.Certificate{}, "", err
	}

	caCertPath := filepath.Join(certDir, "vygrant_ca.pem")
	caKeyPath := filepath.Join(certDir, "vygrant_ca.key")
	leafCertPath := filepath.Join(certDir, "localhost.pem")
	leafKeyPath := filepath.Join(certDir, "localhost.key")

	caCert, caKey, err := ensureCA(caCertPath, caKeyPath)
	if err != nil {
		return tls.Certificate{}, "", err
	}

	leafCert, leafKey, err := ensureLeaf(leafCertPath, leafKeyPath, caCert, caKey)
	if err != nil {
		return tls.Certificate{}, "", err
	}

	pubKeyHex := FormatPublicKeyFromKey(leafKey)
	return leafCert, pubKeyHex, nil
}

// formatPublicKey formats the public key into colon-separated uppercase hex
func FormatPublicKey(b []byte) string {
	hexStr := hex.EncodeToString(b)
	var parts []string
	for i := 0; i < len(hexStr); i += 2 {
		parts = append(parts, strings.ToUpper(hexStr[i:i+2]))
	}
	return strings.Join(parts, ":")
}

func FormatPublicKeyFromKey(priv *ecdsa.PrivateKey) string {
	// Generate public key bytes in uncompressed format (0x04 | X | Y)
	pubkey := append([]byte{0x04}, priv.PublicKey.X.Bytes()...)
	pubkey = append(pubkey, priv.PublicKey.Y.Bytes()...)
	return FormatPublicKey(pubkey)
}

func ensureCA(certPath, keyPath string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	cert, key, err := loadCertAndKey(certPath, keyPath)
	if err == nil && cert.IsCA {
		if err := ensureFilePermissions(certPath, 0o600); err != nil {
			return nil, nil, err
		}
		if err := ensureFilePermissions(keyPath, 0o600); err != nil {
			return nil, nil, err
		}
		return cert, key, nil
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "vygrant local CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	if err := writeFile(certPath, certPEM, 0o600); err != nil {
		return nil, nil, err
	}
	if err := writeFile(keyPath, keyPEM, 0o600); err != nil {
		return nil, nil, err
	}

	parsed, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, err
	}
	return parsed, priv, nil
}

func ensureLeaf(certPath, keyPath string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (tls.Certificate, *ecdsa.PrivateKey, error) {
	const renewWindow = 14 * 24 * time.Hour
	certPEM, keyPEM, err := loadPEMFiles(certPath, keyPath)
	if err == nil {
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err == nil {
			leaf, err := loadFirstCert(certPEM)
			if err == nil && leaf.NotAfter.After(time.Now().Add(renewWindow)) && hasLocalhostSANs(leaf) {
				priv, err := loadECPrivateKeyFromPEM(keyPEM)
				if err == nil {
					if err := ensureFilePermissions(certPath, 0o600); err != nil {
						return tls.Certificate{}, nil, err
					}
					if err := ensureFilePermissions(keyPath, 0o600); err != nil {
						return tls.Certificate{}, nil, err
					}
					return cert, priv, nil
				}
			}
		}
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(90 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	if err := writeFile(certPath, certPEM, 0o600); err != nil {
		return tls.Certificate{}, nil, err
	}
	if err := writeFile(keyPath, keyPEM, 0o600); err != nil {
		return tls.Certificate{}, nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	return cert, priv, nil
}

func loadCertAndKey(certPath, keyPath string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	certPEM, keyPEM, err := loadPEMFiles(certPath, keyPath)
	if err != nil {
		return nil, nil, err
	}
	cert, err := loadFirstCert(certPEM)
	if err != nil {
		return nil, nil, err
	}
	key, err := loadECPrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, nil, err
	}
	return cert, key, nil
}

func loadPEMFiles(certPath, keyPath string) ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

func loadFirstCert(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("invalid cert pem")
	}
	return x509.ParseCertificate(block.Bytes)
}

func loadECPrivateKeyFromPEM(keyPEM []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("invalid key pem")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func hasLocalhostSANs(cert *x509.Certificate) bool {
	hasDNS := false
	for _, name := range cert.DNSNames {
		if strings.EqualFold(name, "localhost") {
			hasDNS = true
			break
		}
	}
	hasV4 := false
	hasV6 := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(net.ParseIP("127.0.0.1")) {
			hasV4 = true
		}
		if ip.Equal(net.ParseIP("::1")) {
			hasV6 = true
		}
	}
	return hasDNS && hasV4 && hasV6
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return err
	}
	return nil
}

func ensureFilePermissions(path string, perm os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != perm {
		return os.Chmod(path, perm)
	}
	return nil
}

func EnsureLocalCA() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	certDir := filepath.Join(home, ".vybr", "vygrant", "certs")
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return "", err
	}

	caCertPath := filepath.Join(certDir, "vygrant_ca.pem")
	caKeyPath := filepath.Join(certDir, "vygrant_ca.key")
	if _, _, err := ensureCA(caCertPath, caKeyPath); err != nil {
		return "", err
	}

	return caCertPath, nil
}
