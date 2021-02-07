package resock

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"os"
)

func DialTLS(address string) (net.Conn, error) {
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	certBytes, err := os.ReadFile("certs/client.pem")
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		CipherSuites:       []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
	}
	return tls.Dial("tcp", address, conf)
}

func ListenTLS(address string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	certBytes, err := os.ReadFile("certs/client.pem")
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	config := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                clientCertPool,
		PreferServerCipherSuites: true,
	}
	return tls.Listen("tcp", address, config)
}
