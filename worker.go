package resock

import (
	"errors"
	"net"
)

type Worker func(conn net.Conn) (net.Conn, error)

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
	host, err := Socks5Connect(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	dstStream, err := NewChacha20Stream(GetCfg().Key, dial)
	dstStream.Write(host)
	return dstStream, err
}

func wsClientWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Connect(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.DialTLS(host.String(), "ws://"+GetCfg().Server)
}

func wssClientWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	host, err := Socks5Connect(conn)
	if err != nil {
		return nil, errors.New("SOCKS error:" + err.Error())
	}
	return ws.DialTLS(host.String(), "wss://"+GetCfg().Server+"/wss")
}
