package resock

import (
	"errors"
	"log"
)

func RunClient() error {
	listener, err := SelectProtocol("tcp", GetCfg().Client)
	defer listener.Close()
	if err != nil {
		return errors.New("listen failed:" + err.Error())
	}
	log.Println("listening on " + GetCfg().Protocol + "://" + GetCfg().Client)
	switch GetCfg().Protocol {
	case "tcp":
		Run(listener, socks5ClientWorker)
	default:
		Run(listener, wsClientWorker)
	}
	return nil
}
