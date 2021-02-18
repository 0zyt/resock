package resock

import (
	"errors"
	"log"
	"net"
	"runtime"
)

func RunServer() error {
	log.Println("listening on " + GetCfg().Protocol + "://" + GetCfg().Server)
	switch GetCfg().Protocol {
	case "tcp":
		listener, err := net.Listen("tcp", GetCfg().Server)
		if err != nil {
			return errors.New("listen failed:" + err.Error())
		}
		RunGroup(runtime.NumCPU(), listener, socks5SWorkers(), true)
	case "wss":
		ws := NewWebsock()
		if err := ws.ListenTLS(GetCfg().Server); err != nil {
			return errors.New("listen failed:" + err.Error())
		}
	default:
		ws := NewWebsock()
		if err := ws.Listen(GetCfg().Server); err != nil {
			return errors.New("listen failed:" + err.Error())
		}
	}
	return nil
}
