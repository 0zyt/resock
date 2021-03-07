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
	cpus := runtime.NumCPU() / 2
	switch GetCfg().Protocol {
	case "tcp":
		RunGroup(cpus, listener, socksLocalPipe(), false)
	case "wss":
		RunGroup(cpus, listener, wsLocalPipe(true), false)
	default:
		RunGroup(cpus, listener, wsLocalPipe(false), false)
	}
	return nil
}
