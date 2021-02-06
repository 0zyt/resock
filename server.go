package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	err        error
	listen     net.Listener
	accept     net.Conn
	acceptChan chan net.Conn = make(chan net.Conn)
)

func main() {
	sockConfig := NewsockConfig("", ":2000")
	listen, err = net.Listen("tcp", sockConfig.localAddr)
	defer listen.Close()
	if err != nil {
		log.Println("listen failed:", err)
	}
	log.Println("listening on " + sockConfig.localAddr + ", to " + sockConfig.remoteAddr)
	acceptor()
}

func acceptor() {
	go worker()
	for {
		accept, err = listen.Accept()
		if err != nil {
			log.Println("accept failed:", err)
			accept.Close()
			continue
		} else {
			acceptChan <- accept
		}
	}
}
func worker() {
	var dstConn net.Conn
	for client := range acceptChan {
		dstConn, err = Socks5Connect(client)
		if err != nil {
			log.Println(err)
			client.Close()
			continue
		}
		go relay(client, dstConn)
	}
}

func relay(src, dst net.Conn) {
	timeout := 5 * time.Second
	defer dst.Close()
	defer src.Close()
	wg := sync.WaitGroup{}
	wg.Add(2)
	forward := func(src, dst net.Conn, wg sync.WaitGroup) {
		io.Copy(src, dst)
		src.SetReadDeadline(time.Now().Add(timeout))
		wg.Done()
	}
	go forward(src, dst, wg)
	go forward(dst, src, wg)
	wg.Wait()
}
