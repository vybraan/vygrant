package daemon

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

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
	Config     *config.Config
	TokenStore storage.TokenStore
	PublicKey  string
	HTTPClient *http.Client
}

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

	auth.LoadedAccounts = cfg.Accounts

	var store storage.TokenStore

	if cfg.PersistTokens {
		home, _ := os.UserHomeDir()
		storePath := path.Join(home, ".vybr/vygrant/tokens.json")
		store = storage.NewFileStore(storePath)
	} else {
		store = storage.NewMemoryStore()
	}

	return &Daemon{
		Config:     cfg,
		TokenStore: store,
	}, nil
}

func (d *Daemon) Start() {
	stopCh := make(chan struct{})
	defer close(stopCh)

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

	if httpEnabled {
		httpServer := &http.Server{
			Handler: handler,
		}
		if httpsEnabled {
			go func() {
				if err := httpServer.Serve(httpListener); err != nil {
					log.Fatalf("http server crashed: %v", err)
				}
			}()
		} else {
			log.Println("oauth2 daemon is running (http only)")
			if err := httpServer.Serve(httpListener); err != nil {
				log.Fatalf("server crashed: %v", err)
			}
			return
		}
	}

	httpsServer := &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	log.Println("oauth2 daemon is running")
	tlsListener := tls.NewListener(httpsListener, httpsServer.TLSConfig)
	if err := httpsServer.Serve(tlsListener); err != nil {
		log.Fatalf("https server crashed: %v", err)
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

func isListenerEnabled(port string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(port))
	return trimmed != "" && trimmed != "none" && trimmed != "off" && trimmed != "disabled"
}

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
