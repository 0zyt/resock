package resock

import (
	"errors"
	"log"
	"net"
)

func RunServer() error {
	log.Println("listening on " + GetCfg().Protocol + "://" + GetCfg().Server)
	switch GetCfg().Protocol {
	case "tcp":
		listener, err := net.Listen("tcp", GetCfg().Server)
		if err != nil {
			return errors.New("listen failed:" + err.Error())
		}
		Run(listener, socks5ServerWorker, true)
	default:
		ws := NewWebsock()
		_, err := ws.Listen(GetCfg().Server)
		if err != nil {
			return errors.New("listen failed:" + err.Error())
		}
	}
	return nil
}
