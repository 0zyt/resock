package resock

import (
	"encoding/binary"
	"errors"
	"net"
)

func Socks5Connect(conn net.Conn) (string, error) {
	if err := socks5Auth(conn); err == nil {
		dstConn, err := socks5Requests(conn)
		if err != nil {
			return "", err
		}
		if err = socks5Reply(conn); err != nil {
			return "", err
		}
		return dstConn, nil
	} else {
		return "", err
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
		return errors.New("version is not 5")
	}
	conn.Read(buf[:nmethods])
	conn.Write(([]byte{5, 0}))
	return nil
}

func socks5Requests(conn net.Conn) (string, error) {
	buf := GetBuf()
	defer PutBuf(buf)
	n, err := conn.Read(buf)
	if err != nil {
		return "", errors.New("socks5 read:" + err.Error())
	}
	atpy := buf[3]
	dstAddr := &net.TCPAddr{}
	switch atpy {
	case 1:
		dstAddr.IP = buf[4 : net.IPv4len+4]
	case 4:
		dstAddr.IP = buf[4 : net.IPv6len+4]
	case 3:
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return "", err
		}
		dstAddr.IP = ipAddr.IP
	}
	dstAddr.Port = int(binary.BigEndian.Uint16(buf[n-2:]))
	return dstAddr.String(), nil
}

func socks5Reply(conn net.Conn) error {
	_, err := conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return err
	}
	return nil
}
