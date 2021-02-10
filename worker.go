package resock

import (
	"errors"
	"net"
)

type Worker func(conn net.Conn) (net.Conn, error)

func socks5ServerWorker(conn net.Conn) (net.Conn, error) {
	dstAddr, err := Socks5Connect(conn)
	if err != nil {
		return nil, errors.New("socks5 handshake " + err.Error())
	}
	return net.Dial("tcp", dstAddr)
}

func socks5ClientWorker(conn net.Conn) (net.Conn, error) {
	return net.Dial("tcp", GetCfg().Server)
}

func wsClientWorker(conn net.Conn) (net.Conn, error) {
	ws := NewWebsock()
	connect, err := Socks5Connect(conn)
	if err != nil {
		return nil, errors.New("socks5 handshake " + err.Error())
	}
	return ws.Dial(connect, "ws://"+GetCfg().Server)
}
