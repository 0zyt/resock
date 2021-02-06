package main

import (
	"log"
	"net"
)

func RunServer() error {
	return Run(getConfig().remoteAddr, ListenTLS)
}

func socks5ServerWorker(accpetChan <-chan net.Conn) {
	for client := range accpetChan {
		dstConn, err := Socks5Connect(client)
		if err != nil {
			log.Println(err)
			client.Close()
			continue
		}
		go relay(client, dstConn)
	}
}
