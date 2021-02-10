package resock

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
)

type websock struct {
	net.Conn
	wsConn     *websocket.Conn
	remoteAddr string
	localAddr  string
	ctx        context.Context
	next       http.Handler
}

var Ws *websock

func NewWebsock() *websock {
	Ws = &websock{ctx: context.Background()}
	return Ws
}

func (w websock) Dial(host, address string) (net.Conn, error) {
	options := &websocket.DialOptions{HTTPHeader: map[string][]string{
		"X-Forwarded-Host": {host},
	}}
	dial, _, err := websocket.Dial(w.ctx, address, options)
	w.wsConn = dial
	if err != nil {
		return nil, err
	}
	w.Conn = websocket.NetConn(w.ctx, dial, websocket.MessageBinary)
	return w.Conn, nil
}

func (w *websock) Listen(address string) (net.Listener, error) {
	w.localAddr = address
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	http.Serve(listen, w)
	return w, nil
}

func (w websock) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	w.remoteAddr = request.Header.Get("X-Forwarded-Host")

	accept, err := websocket.Accept(writer, request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	local := websocket.NetConn(w.ctx, accept, websocket.MessageBinary)
	dial, err := net.Dial("tcp", w.remoteAddr)
	if err != nil {
		log.Println(err)
		return
	}
	go relay(local, dial)
}

func (w *websock) Accept() (net.Conn, error) {
	w.Conn = websocket.NetConn(w.ctx, w.wsConn, websocket.MessageBinary)
	fmt.Println("acc11")
	return w, nil
}

func (w *websock) RemoteAddr() net.Addr {
	return &websockAddr{w.remoteAddr}
}

func (w *websock) Close() error {
	return w.wsConn.Close(websocket.StatusNormalClosure, "")
}

func (w *websock) Addr() net.Addr {
	return &websockAddr{w.localAddr}
}

type websockAddr struct {
	address string
}

func (w *websockAddr) Network() string {
	return "websock"
}

func (w *websockAddr) String() string {
	return w.address
}
