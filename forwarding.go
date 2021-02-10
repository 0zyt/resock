package resock

import (
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

func Run(listen net.Listener, worker Worker) {
	acceptChan := make(chan net.Conn, runtime.NumCPU())
	for {
		accept, err := listen.Accept()
		if err != nil {
			log.Println("accept failed:", err)
			accept.Close()
			continue
		} else {
			acceptChan <- accept
			go process(acceptChan, worker)
		}
	}
}

func process(cChan <-chan net.Conn, worker Worker) {
	for local := range cChan {
		if remote, err := worker(local); err == nil {
			go relay(local, remote)
		} else {
			log.Println(err)
			local.Close()
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
		src.SetDeadline(time.Now().Add(5 * time.Second))
		io.Copy(src, dst)

	}

	wg.Add(2)
	go forward(src, dst)
	go forward(dst, src)
	wg.Wait()

}
