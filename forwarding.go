package resock

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func SelectProtocol(network, address string) (net.Listener, error) {
	switch network {
	case "tcp":
		return net.Listen(network, address)
	case "tls":
		return ListenTLS(address)
	default:
		return nil, errors.New("unsupported protocol")
	}
}

func Run(listener net.Listener, worker func(accpetChan <-chan net.Conn)) error {

	acceptor(listener, worker)
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
		buf := bufferPoolGet()
		defer wg.Done()
		defer bufferPoolPut(buf)
		src.SetDeadline(time.Now().Add(10 * time.Second))
		io.Copy(src, dst)

	}

	wg.Add(2)
	go forward(src, dst)
	go forward(dst, src)
	wg.Wait()

}
