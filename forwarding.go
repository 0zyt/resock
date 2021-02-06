package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func Run(lsIP string, listen func(network, address string) (net.Listener, error)) error {
	listener, err := listen("tcp", lsIP)
	defer listener.Close()
	if err != nil {
		log.Println("listen failed:", err)
		return err
	}
	log.Println("listening on " + lsIP)
	acceptor(listener, socks5ServerWorker)
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
	defer src.Close()
	defer dst.Close()
	wg := sync.WaitGroup{}

	forward := func(src, dst net.Conn) {
		defer wg.Done()
		src.SetDeadline(time.Now().Add(10 * time.Second))
		io.Copy(src, dst)
	}

	wg.Add(2)
	go forward(src, dst)
	go forward(dst, src)
	wg.Wait()

}
