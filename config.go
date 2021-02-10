package resock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"sync"
)

type config struct {
	Server   string
	Client   string
	Protocol string
	Username string
	Password string
}

var (
	cfg *config = &config{}
	so  sync.Once
)

func GetCfg() *config {
	so.Do(func() {
		if file, err := os.ReadFile("cfg.json"); err == nil {
			json.Unmarshal(file, cfg)
			fmt.Println("Init config")
		}
	})
	return cfg
}

func GenCfg() {
	b, _ := json.MarshalIndent(&config{Server: "", Client: "2", Protocol: "3", Username: "4", Password: "5"}, " ", " ")
	os.WriteFile("cfg.json", b, fs.ModePerm)
}

func SelectProtocol(network, address string) (net.Listener, error) {
	switch network {
	case "tcp":
		return net.Listen(network, address)
	case "tls":
		return ListenTLS(address)
	case "ws":
		return NewWebsock().Listen(address)
	default:
		return nil, errors.New("unsupported Protocol")
	}
}
