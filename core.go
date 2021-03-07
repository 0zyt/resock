package resock

import (
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

var globalCh chan net.Conn

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

func RunGroup(nums int, listen net.Listener, workers *Pipeline, isServer bool) {
	globalCh = make(chan net.Conn, 65535)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go acceptor(listen, workers, isServer, wg)
		go process(workers)
	}
}

func dispatch(localCh chan<- net.Conn) {
	for conn := range globalCh {
		localCh <- conn
	}
}

func acceptor(listen net.Listener, pipe *Pipeline, isSrv bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		accept, err := listen.Accept()
		if err != nil {
			log.Println("accept failed:", err)
			accept.Close()
			continue
		} else {
			accept.(*net.TCPConn).SetLinger(0)
			if isSrv {
				var err error
				accept, err = pipe.Filter(accept, isSrv)
				if err != nil {
					log.Println(err)
					continue
				}
			}
			globalCh <- accept
		}
	}
}

func process(pipe *Pipeline) {
	localCh := make(chan net.Conn, runtime.NumCPU())
	go dispatch(localCh)
	//如果收到请求是经过加密或者其他操作，需要先统一在上面的流水线里对Conn进行相关的转换,保证这里读到的是正确是数据
	for local := range localCh {
		if remote, err := pipe.Filter(local, false); err == nil {
			go relay(local, remote)
		} else {
			log.Println(err)
			//如果出现错误，把这次的请求全丢掉，清空缓冲区
			io.Copy(io.Discard, local)
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
		src.SetReadDeadline(time.Now().Add(5 * time.Second))
		io.CopyBuffer(src, dst, GetBuf())

	}

	wg.Add(2)
	go forward(src, dst)
	go forward(dst, src)
	wg.Wait()

}
