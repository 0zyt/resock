package resock

import (
	"errors"
	"net"
)

type Worker func(conn net.Conn) (net.Conn, error)

type Workers struct {
	in   []Worker
	out  []Worker
	host Addr
}

func (p *Workers) Filter(conn net.Conn, isSrv bool) (net.Conn, error) {
	var err error
	var pipe = p.out
	if isSrv {
		pipe = p.in
	}
	for _, worker := range pipe {
		conn, err = worker(conn)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (p *Workers) AddOut(w ...Worker) {
	p.out = append(p.out, w...)
}

func (p *Workers) AddIn(w ...Worker) {
	p.in = append(p.in, w...)
}

func (p Workers) Add(in []Worker, out []Worker) {
	if in != nil {
		p.AddIn(in...)
	}
	if out != nil {
		p.AddOut(out...)
	}
}

func socks5CWorkers() *Workers {
	p := &Workers{}
	logical1Worker := func(conn net.Conn) (net.Conn, error) {
		host, err := Socks5Handshake(conn)
		if err != nil {
			return nil, errors.New("SOCKS error:" + err.Error())
		}
		p.host = host
		return conn, nil
	}
	logical2Worker := func(conn net.Conn) (net.Conn, error) {
		_, err := conn.Write(p.host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	p.AddOut(logical1Worker, basicTCPToSrvWorker, chacha20Worker, logical2Worker)
	return p
}

func socks5SWorkers() *Workers {
	p := &Workers{}
	p.AddIn(chacha20Worker)
	p.AddOut(func(conn net.Conn) (net.Conn, error) {
		buf := GetBuf()
		defer PutBuf(buf)
		addr, err := readAddr(buf, conn)
		if err != nil {
			return nil, errors.New("SOCKS error:" + err.Error())
		}
		p.host = addr
		return conn, nil
	},
		func(conn net.Conn) (net.Conn, error) {
			return net.Dial("tcp", p.host.String())
		})
	return p
}

func chacha20Worker(conn net.Conn) (net.Conn, error) {
	return NewChacha20Stream(GetCfg().Key, conn)
}

func basicTCPToSrvWorker(conn net.Conn) (net.Conn, error) {
	return net.Dial("tcp", GetCfg().Server)
}

func socks5ServerWorker(conn net.Conn) (net.Conn, error) {
	buf := GetBuf()
	defer PutBuf(buf)
	addr, err := readAddr(buf, conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return net.Dial("tcp", addr.String())
}

func socks5ClientWorker(conn net.Conn) (net.Conn, error) {
	dial, err := net.Dial("tcp", GetCfg().Server)
	if err != nil {
		return nil, err
	}
	host, err := Socks5Handshake(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	dstStream, err := NewChacha20Stream(GetCfg().Key, dial)
	if err != nil {
		return nil, err
	}
	_, err = dstStream.Write(host)
	if err != nil {
		return nil, err
	}
	return dstStream, nil
}

func wsClientWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Handshake(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.Dial(host.String(), "ws://"+GetCfg().Server)
}

func wssClientWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Handshake(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.DialTLS(host.String(), "wss://"+GetCfg().Server+"/wss")
}
