package daemon

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/vybraan/vygrant/internal/api"
	"github.com/vybraan/vygrant/internal/auth"
	"github.com/vybraan/vygrant/internal/certgen"
	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
)

const (
	VYGRANT_CONFIG = ".config/vybr/vygrant.toml"
)

type Daemon struct {
	Config          *config.Config
	TokenStore      storage.TokenStore
	PublicKey       string
	HTTPClient      *http.Client
	LegacyMigration string
}

// NewDaemon creates a Daemon by loading configuration and initializing token storage.
//
// It loads configuration from the path specified by the VYGRANT_CONFIG environment
// variable or from the default user config path (~/.config/vybr/vygrant.toml). If
// configuration loading fails, an error is returned. The function populates
// auth.LoadedAccounts from the loaded configuration and selects the token store:
// when cfg.PersistTokens is true it prefers the OS keyring for refresh tokens (with
// an in-memory access-token cache). When cfg.PersistTokens is false, or when a keyring
// is unavailable, it uses an in-memory store only. The returned Daemon has Config and
// TokenStore initialized.
func NewDaemon() (*Daemon, error) {
	confPath := os.Getenv("VYGRANT_CONFIG")
	if confPath == "" {
		home, _ := os.UserHomeDir()
		confPath = path.Join(home, VYGRANT_CONFIG)
	}

	cfg, err := config.LoadConfig(confPath)
	if err != nil {
		return nil, err
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	auth.LoadedAccounts = cfg.Accounts

	var store storage.TokenStore
	home, _ := os.UserHomeDir()
	legacyTokenPath := path.Join(home, ".vybr/vygrant/tokens.json")

	if cfg.PersistTokens {
		if storage.KeyringAvailable("") {
			refreshStore := storage.NewKeyringStore("")
			splitStore := storage.NewSplitStore(refreshStore)
			if migrated, backupPath, err := migrateLegacyTokens(legacyTokenPath, splitStore); err != nil {
				log.Printf("warning: failed to migrate legacy tokens: %v", err)
			} else if migrated {
				log.Printf("migrated legacy tokens to keyring; backed up file to %s", backupPath)
				store = splitStore
				return &Daemon{
					Config:          cfg,
					TokenStore:      store,
					LegacyMigration: fmt.Sprintf("migrated legacy file to keyring (backup: %s)", backupPath),
				}, nil
			}
			store = splitStore
		} else {
			if fileExists(legacyTokenPath) {
				log.Println("warning: keyring unavailable; using legacy file token store")
				store = storage.NewFileStore(legacyTokenPath)
			} else {
				log.Println("warning: token persistence unavailable; falling back to in-memory token store")
				store = storage.NewMemoryStore()
			}
		}
	} else {
		store = storage.NewMemoryStore()
	}

	return &Daemon{
		Config:     cfg,
		TokenStore: store,
	}, nil
}

func (d *Daemon) Start() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stopCh := make(chan struct{})

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	d.HTTPClient = httpClient

	go StartBackgroundTasks(d.Config, d.TokenStore, d.HTTPClient, stopCh)

	httpsEnabled := isListenerEnabled(d.Config.HTTPSListen)
	httpEnabled := isListenerEnabled(d.Config.HTTPListen)
	if !httpsEnabled && !httpEnabled {
		log.Fatal("no HTTP or HTTPS listener configured")
	}

	var cert tls.Certificate
	if httpsEnabled {
		publicKey := ""
		var err error
		cert, publicKey, err = certgen.GenerateSelfSignedCert()
		if err != nil {
			log.Fatalf("tls setup failed: %v", err)
		}
		d.PublicKey = publicKey
	}

	handler := api.Router(&d.TokenStore, d.HTTPClient)

	httpAddr := "localhost:" + d.Config.HTTPListen
	httpsAddr := "localhost:" + d.Config.HTTPSListen

	var httpListener net.Listener
	var httpsListener net.Listener
	var err error

	if httpEnabled {
		httpListener, err = net.Listen("tcp", httpAddr)
		if err != nil {
			log.Fatalf("http listener failed: %v", err)
		}
	}

	if httpsEnabled {
		httpsListener, err = net.Listen("tcp", httpsAddr)
		if err != nil {
			if httpListener != nil {
				httpListener.Close()
			}
			log.Fatalf("https listener failed: %v", err)
		}
	}

	socketPath, err := ensureSocketAvailable()
	if err != nil {
		if httpListener != nil {
			httpListener.Close()
		}
		if httpsListener != nil {
			httpsListener.Close()
		}
		log.Fatal(err)
	}

	socketListener, err := net.Listen("unix", socketPath)
	if err != nil {
		if httpListener != nil {
			httpListener.Close()
		}
		if httpsListener != nil {
			httpsListener.Close()
		}
		log.Fatalf("socket listener failed: %v", err)
	}
	defer func() {
		socketListener.Close()
		os.Remove(socketPath)
	}()
	go d.handleConnections(socketListener)

	errCh := make(chan error, 2)

	var httpServer *http.Server
	if httpEnabled {
		httpServer = &http.Server{Handler: handler}
		go func() {
			if err := httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("http server crashed: %w", err)
			}
		}()
	}

	httpsServer := &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	if httpsEnabled {
		log.Println("oauth2 daemon is running")
		tlsListener := tls.NewListener(httpsListener, httpsServer.TLSConfig)
		go func() {
			if err := httpsServer.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("https server crashed: %w", err)
			}
		}()
	} else {
		log.Println("oauth2 daemon is running (http only)")
	}

	select {
	case err := <-errCh:
		log.Fatal(err)
	case <-ctx.Done():
		log.Println("shutting down daemon")
	}

	close(stopCh)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}
	if httpsEnabled {
		_ = httpsServer.Shutdown(shutdownCtx)
	}
	if httpListener != nil {
		_ = httpListener.Close()
	}
	if httpsListener != nil {
		_ = httpsListener.Close()
	}
}

func (d *Daemon) handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go d.handle(conn)
	}

}

func (d *Daemon) handle(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()
		d.HandleCommand(conn, input)
	}
}

// isListenerEnabled reports whether the given listener port string enables a listener.
// It treats an empty string or the values "none", "off", and "disabled" (case-insensitive, with surrounding whitespace ignored) as disabled; all other values are considered enabled.
func isListenerEnabled(port string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(port))
	return trimmed != "" && trimmed != "none" && trimmed != "off" && trimmed != "disabled"
}

// ensureSocketAvailable verifies the application's UNIX socket path is usable and returns it.
//
// It returns an error if the socket path cannot be determined, if the path exists but is not a UNIX socket,
// or if another process is already listening on the socket. If a stale socket file exists and is removable,
// it removes that file and returns the path.
func ensureSocketAvailable() (string, error) {
	socketPath := SocketPath()
	if socketPath == "" {
		return "", fmt.Errorf("could not determine socket path")
	}

	info, err := os.Stat(socketPath)
	if err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return "", fmt.Errorf("socket path exists and is not a socket: %s", socketPath)
		}
		if conn, err := net.Dial("unix", socketPath); err == nil {
			conn.Close()
			return "", fmt.Errorf("daemon already running on %s", socketPath)
		}
		if err := os.Remove(socketPath); err != nil {
			return "", fmt.Errorf("failed to remove stale socket: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat socket: %v", err)
	}

	return socketPath, nil
}

func migrateLegacyTokens(path string, store storage.TokenStore) (bool, string, error) {
	if !fileExists(path) {
		return false, "", nil
	}

	legacyStore := storage.NewFileStore(path)
	accounts := legacyStore.ListAccounts()
	if len(accounts) == 0 {
		return false, "", nil
	}

	for _, account := range accounts {
		token, err := legacyStore.Get(account)
		if err != nil {
			continue
		}
		if token == nil {
			continue
		}
		if token.AccessToken == "" && token.RefreshToken == "" {
			continue
		}
		if err := store.Set(account, token); err != nil {
			return false, "", err
		}
	}

	backupPath := path + ".bak"
	if fileExists(backupPath) {
		backupPath = fmt.Sprintf("%s.%d", backupPath, time.Now().Unix())
	}
	if err := os.Rename(path, backupPath); err != nil {
		return false, "", err
	}

	return true, backupPath, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func validateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if len(cfg.Accounts) == 0 {
		return nil
	}

	httpsEnabled := isListenerEnabled(cfg.HTTPSListen)
	httpEnabled := isListenerEnabled(cfg.HTTPListen)

	for name, acct := range cfg.Accounts {
		if acct == nil {
			return fmt.Errorf("account %q is nil", name)
		}
		if acct.AuthURI == "" || acct.TokenURI == "" || acct.RedirectURI == "" || acct.ClientID == "" {
			return fmt.Errorf("account %q is missing required fields", name)
		}
		if err := validateURL(acct.AuthURI, "auth_uri", name); err != nil {
			return err
		}
		if err := validateURL(acct.TokenURI, "token_uri", name); err != nil {
			return err
		}
		if err := validateURL(acct.RedirectURI, "redirect_uri", name); err != nil {
			return err
		}

		redirectURL, _ := url.Parse(acct.RedirectURI)
		switch redirectURL.Scheme {
		case "https":
			if !httpsEnabled {
				return fmt.Errorf("account %q redirect_uri is https but https_listen is disabled", name)
			}
		case "http":
			if !httpEnabled {
				return fmt.Errorf("account %q redirect_uri is http but http_listen is disabled", name)
			}
		default:
			return fmt.Errorf("account %q redirect_uri must be http or https", name)
		}
	}

	return nil
}

func validateURL(rawURL, field, account string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("account %q has invalid %s: %v", account, field, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("account %q %s must be http or https", account, field)
	}
	if parsed.Host == "" {
		return fmt.Errorf("account %q %s missing host", account, field)
	}
	return nil
}
