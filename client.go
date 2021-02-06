package main

import (
	"log"
	"net"
)

func RunClient() error {
	return Run(getConfig().localAddr, net.Listen)
}

func socks5Clientworker(accpetChan <-chan net.Conn) {
	for local := range accpetChan {
		remote, err := DialTLS(getConfig().remoteAddr)
		if err != nil {
			log.Println(err)
			local.Close()
			continue
		}
		go relay(local, remote)
	}
}
