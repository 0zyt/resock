package resock

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

var acChan chan net.Conn

var bufPool sync.Pool

func init() {
	bufPool.New = func() interface{} {
		//refer Linux cat /proc/sys/net/core/optmem_max
		return make([]byte, 20480)
	}
}
func GetBuf() []byte {
	return bufPool.Get().([]byte)
}
func PutBuf(b []byte) {
	bufPool.Put(b)
}

func RunGroup(nums int, listen net.Listener, worker Worker, isServer bool) {
	acChan = make(chan net.Conn, runtime.NumCPU())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go acceptor(listen, worker, isServer, wg)
	}
}

func acceptor(listen net.Listener, worker Worker, isServer bool, wg sync.WaitGroup) {
	defer wg.Done()
	for {
		accept, err := listen.Accept()
		if err != nil {
			log.Println("accept failed:", err)
			accept.Close()
			continue
		} else {
			if isServer {
				var err error
				accept, err = NewChacha20Stream(GetCfg().Key, accept)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			acChan <- accept
			go process(acChan, worker)
		}
	}
}

func process(acChan <-chan net.Conn, worker Worker) {
	for local := range acChan {
		if remote, err := worker(local); err == nil {
			go relay(local, remote)
		} else {
			log.Println(err)
			buf := GetBuf()
			defer PutBuf(buf)
			_, err = io.CopyBuffer(ioutil.Discard, local, buf)
			if err != nil {
				log.Println(err)
			}
			local.Close()
		}
	}
}

func relay(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	wg := sync.WaitGroup{}

	forward := func(src, dst net.Conn) {
		buf := GetBuf()
		defer wg.Done()
		defer PutBuf(buf)
		src.SetDeadline(time.Now().Add(5 * time.Second))
		//io.Copy(src, dst)
		io.CopyBuffer(src, dst, GetBuf())

	}

	wg.Add(2)
	go forward(src, dst)
	go forward(dst, src)
	wg.Wait()

}
