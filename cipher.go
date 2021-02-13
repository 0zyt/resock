package resock

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"nhooyr.io/websocket"
	"os"
	"time"
)

func getClientCert() *tls.Config {
	cert, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
	if err != nil {
		log.Println(err)
		return nil
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
	return &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		ServerName:         GetCfg().SNI,
	}
}

func DialTLS(address string) (net.Conn, error) {
	return tls.Dial("tcp", address, getClientCert())
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

func (w *websock) ListenTLS(address string) (net.Listener, error) {
	w.localAddr = address
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		parse, _ := url.Parse(GetCfg().SNI)
		httputil.NewSingleHostReverseProxy(parse).ServeHTTP(writer, request)
	})
	http.Handle("/wss", w)
	http.ListenAndServeTLS(address, "certs/server.pem", "certs/server.key", nil)
	return w, nil
}

func (w websock) DialTLS(host, address string) (net.Conn, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: getClientCert(),
		},
	}
	options := &websocket.DialOptions{
		HTTPClient: client,
		HTTPHeader: map[string][]string{"X-Forwarded-Host": {host}}}
	dial, _, err := websocket.Dial(w.ctx, address, options)
	w.wsConn = dial
	if err != nil {
		return nil, err
	}
	w.Conn = websocket.NetConn(w.ctx, dial, websocket.MessageBinary)
	return w.Conn, nil
}

func GenKey(key string) []byte {
	h := sha3.New256()
	h.Write([]byte(key))
	return h.Sum([]byte{})
}

type Chacha20Stream struct {
	key     []byte
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	conn    net.Conn
}

func NewChacha20Stream(key []byte, conn net.Conn) (*Chacha20Stream, error) {
	s := &Chacha20Stream{
		key:  key,
		conn: conn,
	}

	nonce := make([]byte, chacha20.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	var err error
	s.encoder, err = chacha20.NewUnauthenticatedCipher(s.key, nonce)
	if err != nil {
		return nil, err
	}

	if n, err := s.conn.Write(nonce); err != nil || n != len(nonce) {
		return nil, errors.New("write nonce failed: " + err.Error())
	}
	return s, nil
}

func (s *Chacha20Stream) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Chacha20Stream) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Chacha20Stream) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *Chacha20Stream) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Chacha20Stream) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

func (s *Chacha20Stream) Read(p []byte) (int, error) {
	if s.decoder == nil {
		nonce := make([]byte, chacha20.NonceSizeX)
		if n, err := io.ReadAtLeast(s.conn, nonce, len(nonce)); err != nil || n != len(nonce) {
			return n, errors.New("can't read nonce from stream: " + err.Error())
		}
		decoder, err := chacha20.NewUnauthenticatedCipher(s.key, nonce)
		if err != nil {
			return 0, errors.New("generate decoder failed: " + err.Error())
		}
		s.decoder = decoder
	}

	n, err := s.conn.Read(p)
	if err != nil || n == 0 {
		return n, err
	}

	dst := make([]byte, n)
	pn := p[:n]
	s.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, nil
}

func (s *Chacha20Stream) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	s.encoder.XORKeyStream(dst, p)
	return s.conn.Write(dst)
}

func (s *Chacha20Stream) Close() error {
	return s.conn.Close()
}
