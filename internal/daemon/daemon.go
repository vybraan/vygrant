package daemon

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/vybraan/vygrant/internal/api"
	"github.com/vybraan/vygrant/internal/auth"
	"github.com/vybraan/vygrant/internal/certgen"
	"github.com/vybraan/vygrant/internal/config"
	"github.com/vybraan/vygrant/internal/storage"
)

const (
	VYGRANT_CONFIG = ".config/vybr/vygrant.toml"
	SOCK           = "/tmp/vygrant.sock"
)

type Daemon struct {
	Config     *config.Config
	TokenStore storage.TokenStore
	PublicKey  string
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

	go StartBackgroundTasks(d.Config, d.TokenStore, stopCh)

	cert, publicKey, err := certgen.GenerateSelfSignedCert()
	if err != nil {
		log.Fatalf("tls setup failed: %v", err)
	}
	d.PublicKey = publicKey

	os.Remove(SOCK)
	listener, err := net.Listen("unix", SOCK)
	if err != nil {
		log.Fatalf("listener failed: %v", err)
	}
	defer func() {
		listener.Close()
		os.Remove(SOCK)
	}()
	go d.handleConnections(listener)

	server := &http.Server{
		Addr:    "localhost:" + d.Config.HTTPSListen,
		Handler: api.Router(&d.TokenStore),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	log.Println("oauth2 daemon is running")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server crashed: %v", err)
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
