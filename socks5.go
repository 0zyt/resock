package resock

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

func Socks5Handshake(conn net.Conn) (Addr, error) {
	if err := socks5Auth(conn); err == nil {
		dstConn, err := socks5Requests(conn)
		if err != nil {
			return nil, err
		}
		if err = socks5Reply(conn); err != nil {
			return nil, err
		}
		return dstConn, nil
	} else {
		return nil, err
	}
}

func socks5Auth(conn net.Conn) error {
	buf := GetBuf()
	defer PutBuf(buf)
	n, err := conn.Read(buf[:2])
	if err != nil || n != 2 {
		return err
	}
	ver, nmethods := buf[0], buf[1]
	if ver != 5 {
		return errors.New("Version Error")
	}
	conn.Read(buf[:nmethods])
	conn.Write(([]byte{5, 0}))
	return nil
}

type Addr []byte

func (buf Addr) String() string {
	dstAddr := &net.TCPAddr{}
	n := len(buf)

	switch buf[0] {
	case 1:
		dstAddr.IP = []byte(buf[1 : net.IPv4len+1])
	case 4:
		dstAddr.IP = []byte(buf[1 : net.IPv6len+1])
	case 3:
		ipAddr, _ := net.ResolveIPAddr("ip", string(buf[2:n-2]))
		dstAddr.IP = ipAddr.IP
	}
	dstAddr.Port = int(binary.BigEndian.Uint16(buf[n-2:]))
	return dstAddr.String()
}

func socks5Requests(conn net.Conn) (Addr, error) {
	buf := GetBuf()
	defer PutBuf(buf)
	_, err := conn.Read(buf[:3])
	if err != nil {
		return nil, err
	}
	addr, err := readAddr(buf, conn)
	return addr, nil
}

func readAddr(buf []byte, r io.Reader) (Addr, error) {
	_, err := r.Read(buf[:1])
	if err != nil {
		return nil, err
	}
	switch buf[0] {
	case 1:
		_, err = r.Read(buf[1 : 1+net.IPv4len+2])
		return buf[:1+net.IPv4len+2], err
	case 4:
		_, err = r.Read(buf[1 : 1+net.IPv6len+2])
		return buf[:1+net.IPv6len+2], err
	case 3:
		_, err = r.Read(buf[1:2])
		if err != nil {
			return nil, err
		}
		_, err = r.Read(buf[2 : 2+int(buf[1])+2])
		return buf[:1+1+int(buf[1])+2], err
	}
	return nil, errors.New("Error Address")
}

func socks5Reply(conn net.Conn) error {
	_, err := conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return err
	}
	return nil
}
