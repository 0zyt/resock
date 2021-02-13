package resock

import (
	"errors"
	"log"
	"net"
)

func RunClient() error {
	listener, err := net.Listen("tcp", GetCfg().Client)
	defer listener.Close()
	if err != nil {
		return errors.New("listen failed:" + err.Error())
	}
	log.Println("listening on " + GetCfg().Protocol + "://" + GetCfg().Client)
	switch GetCfg().Protocol {
	case "tcp":
		Run(listener, socks5ClientWorker, false)
	case "wss":
		Run(listener, wssClientWorker, false)
	default:
		Run(listener, wsClientWorker, false)
	}
	return nil
}
