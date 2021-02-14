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
		RunGroup(runtime.NumCPU(), listener, socks5ServerWorker, true)
	case "wss":
		ws := NewWebsock()
		_, err := ws.ListenTLS(GetCfg().Server)
		if err != nil {
			return errors.New("listen failed:" + err.Error())
		}
	default:
		ws := NewWebsock()
		_, err := ws.Listen(GetCfg().Server)
		if err != nil {
			return errors.New("listen failed:" + err.Error())
		}
	}
	return nil
}
