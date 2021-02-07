package resock

import (
	"log"
	"net"
)

func RunServer() error {
	listener, err := SelectProtocol("kcp", getConfig().remoteAddr)
	defer listener.Close()
	if err != nil {
		log.Println("listen failed:", err)
		return err
	}
	log.Println("listening on " + getConfig().remoteAddr)
	return Run(listener, socks5ServerWorker)
}

func socks5ServerWorker(accpetChan <-chan net.Conn) {
	for client := range accpetChan {
		dstConn, err := Socks5Connect(client)
		if err != nil {
			log.Println("Server:" + err.Error())
			client.Close()
			continue
		}
		go relay(client, dstConn)
	}
}
