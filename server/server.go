package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"playproxy/confer"
	"playproxy/util"
	"sync"
	"time"
)

type Server struct {
	confer *confer.Confer

	authRequired bool

	authCredentials [][]byte // slice with base64-encoded credentials

	dialContext func(ctx context.Context, network, address string) (net.Conn, error)

	bufferPool sync.Pool
}

func newServer(confer *confer.Confer) *Server {
	dialer := &net.Dialer{
		Timeout:   time.Second * 20,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	server := &Server{
		confer:          confer,
		dialContext:     dialer.DialContext,
		authCredentials: make([][]byte, 0),
		bufferPool:      sync.Pool{New: func() interface{} { return make([]byte, 0, 32*1024) }},
	}
	// base64-encode credentials
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(confer.Username)+1+len(confer.Password)))
	base64.StdEncoding.Encode(buf, []byte(confer.Username+":"+confer.Password))
	server.authCredentials = append(server.authCredentials, buf)
	return server
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	// 显示信息
	if s.onAccept(ctx, res, req) {
		return
	}
	// 判断用户鉴权
	authErr := s.onAuth(ctx, req)
	if len(s.confer.ProbeResistDomain) > 0 && util.StripPort(req.Host) == s.confer.ProbeResistDomain {
		// 如果网页输入的是鉴权的地址，返回鉴权结果页面
		serveHiddenPage(res, authErr)
		return
	}
	if authErr != nil {
		// 代表鉴权失败，且访问的不是鉴权地址，反代到欺骗地址
		s.proxyCheatAddress(res, req)
		return
	}
	s.onRequest(ctx, res, req)
}

func ListenAndServe() error {
	config := confer.GlobalConfig()
	handler := newServer(config)
	httpServer := &http.Server{
		Handler: handler,
	}
	l, err := net.Listen("tcp4", config.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("start", config.ListenAddress)
	return httpServer.ServeTLS(l, config.CertFile, config.KeyFile)
}
