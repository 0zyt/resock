package resock

import (
	"context"
	"errors"
	"net"
)

type Worker func(conn net.Conn) (net.Conn, error)

type Pipeline struct {
	in  []Worker
	out []Worker
	ctx context.Context
}

func (p *Pipeline) Filter(conn net.Conn, isSrv bool) (net.Conn, error) {
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

func (p *Pipeline) AddOut(w ...Worker) {
	p.out = append(p.out, w...)
}

func (p *Pipeline) AddIn(w ...Worker) {
	p.in = append(p.in, w...)
}

func (p *Pipeline) Add(in []Worker, out []Worker) {
	if in != nil {
		p.AddIn(in...)
	}
	if out != nil {
		p.AddOut(out...)
	}
}

func socksLocalPipe() *Pipeline {
	p := &Pipeline{}
	logical1Worker := func(conn net.Conn) (net.Conn, error) {
		host, err := Socks5Handshake(conn)
		if err != nil {
			return nil, errors.New("SOCKS error:" + err.Error())
		}
		p.ctx = context.WithValue(context.Background(), "host", host)
		return conn, nil
	}
	logical2Worker := func(conn net.Conn) (net.Conn, error) {
		host := p.ctx.Value("host").(Addr)
		_, err := conn.Write(host)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	p.AddOut(logical1Worker, basicTCPToSrvWorker, chacha20Worker, logical2Worker)
	return p
}

func socksSrvPipe() *Pipeline {
	p := &Pipeline{}
	p.AddIn(chacha20Worker)
	p.AddOut(func(conn net.Conn) (net.Conn, error) {
		buf := GetBuf()
		defer PutBuf(buf)
		addr, err := readAddr(buf, conn)

		if err != nil {
			return nil, errors.New("SOCKS error:" + err.Error())
		}

		p.ctx = context.WithValue(context.Background(), "host", addr)
		return conn, nil
	},
		func(conn net.Conn) (net.Conn, error) {
			host := p.ctx.Value("host").(Addr)
			if "" == host.String() {
				return nil, errors.New("dst addr parse error")
			}
			return net.Dial("tcp", host.String())
		})
	return p
}

func wsLocalPipe(isTLS bool) *Pipeline {
	p := &Pipeline{}
	if isTLS {
		p.AddOut(wssLocalWorker)
	} else {
		p.AddOut(wsLocalWorker)
	}
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

func wsLocalWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Handshake(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.Dial(host.String(), "ws://"+GetCfg().Server)
}

func wssLocalWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Handshake(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.DialTLS(host.String(), "wss://"+GetCfg().Server+"/wss")
}
