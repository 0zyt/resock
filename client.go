package resock

import (
	"errors"
	"log"
	"net"
	"runtime"
)

func RunClient() error {
	listener, err := net.Listen("tcp", GetCfg().Client)
	if err != nil {
		return errors.New("listen failed:" + err.Error())
	}
	defer listener.Close()
	log.Println("listening on " + GetCfg().Protocol + "://" + GetCfg().Client)
	switch GetCfg().Protocol {
	case "tcp":
		RunGroup(runtime.NumCPU(), listener, socksLocalPipe(), false)
	case "wss":
		RunGroup(runtime.NumCPU(), listener, wsLocalPipe(true), false)
	default:
		RunGroup(runtime.NumCPU(), listener, wsLocalPipe(false), false)
	}
	return nil
}
