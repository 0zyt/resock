package resock

import (
	"log"
	"net"
)

func RunClient() error {
	listener, err := SelectProtocol("tcp", getConfig().localAddr)
	defer listener.Close()
	if err != nil {
		log.Println("listen failed:", err)
		return err
	}
	log.Println("listening on " + getConfig().localAddr)
	return Run(listener, socks5Clientworker)
}

func socks5Clientworker(accpetChan <-chan net.Conn) {
	for local := range accpetChan {
		remote, err := net.Dial("tcp", getConfig().remoteAddr)
		//remote, err := DialTLS(getConfig().remoteAddr)
		if err != nil {
			log.Println("Client:" + err.Error())
			local.Close()
			continue
		}
		go relay(local, remote)
	}
}
