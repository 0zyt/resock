package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func Run(lsIP string) error {
	listen, err := net.Listen("tcp", lsIP)
	defer listen.Close()
	if err != nil {
		log.Println("listen failed:", err)
		return err
	}
	log.Println("listening on " + lsIP)
	acceptor(listen, socks5ServerWorker)
	return nil
}

func acceptor(listen net.Listener, worker func(accpetChan <-chan net.Conn)) {
	acceptChan := make(chan net.Conn)
	go worker(acceptChan)
	for {
		accept, err := listen.Accept()
		if err != nil {
			log.Println("accept failed:", err)
			accept.Close()
			continue
		} else {
			acceptChan <- accept
		}
	}
}

func relay(src, dst net.Conn) {
	defer dst.Close()
	defer src.Close()
	wg := sync.WaitGroup{}
	wg.Add(2)
	forward := func(src, dst net.Conn, wg sync.WaitGroup) {
		io.Copy(src, dst)
		src.SetReadDeadline(time.Now().Add(5 * time.Second))
		wg.Done()
	}
	go forward(src, dst, wg)
	go forward(dst, src, wg)
	wg.Wait()
}
