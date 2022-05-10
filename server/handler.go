package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) onAccept(ctx context.Context, res http.ResponseWriter, req *http.Request) bool {
	defer func() {
		if err, ok := recover().(error); ok {
			log.Printf("ERR: %s: %s [%s]", "Accept", ErrPanic, err)
		}
	}()
	// 显示信息，匹配路由之后不做代理
	if req.Method == "GET" && !req.URL.IsAbs() && req.URL.Path == "/info" {
		res.Write([]byte("This is ShadowProxy."))
		return true
	}
	return false
}

func (s *Server) onAuth(ctx context.Context, req *http.Request) (err error) {
	defer func() {
		if err, ok := recover().(error); ok {
			log.Printf("ERR: %s: %s [%s]", "Auth", ErrPanic, err)
		}
	}()
	// 执行鉴权逻辑
	pa := strings.Split(req.Header.Get("Proxy-Authorization"), " ")
	if len(pa) != 2 {
		err = errors.New("Proxy-Authorization is required! Expected format: <type> <credentials>")
		log.Printf("ERR: %s: %s [%s]", "Auth", ErrPanic, err)
		return
	}
	if strings.ToLower(pa[0]) != "basic" {
		err = errors.New("auth type is not supported")
		log.Printf("ERR: %s: %s [%s]", "Auth", ErrPanic, err)
		return
	}
	// 判断用户密码是否正确
	for _, cred := range s.authCredentials {
		if subtle.ConstantTimeCompare(cred, []byte(pa[1])) == 1 {
			// Please do not consider this to be timing-attack-safe code. Simple equality is almost
			// mindlessly substituted with constant time algo and there ARE known issues with this code,
			// e.g. size of smallest credentials is guessable. TODO: protect from all the attacks! Hash?
			return nil
		}
	}
	return errors.New("invalid credentials")
}

func (s *Server) onRequest(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok {
			log.Printf("ERR: %s: %s [%s]", "Request", ErrPanic, err)
		}
	}()
	// 不是http1，也不是http2，返回未知错误
	if req.ProtoMajor != 1 && req.ProtoMajor != 2 {
		err := errors.New("Unsupported HTTP major version: " + strconv.Itoa(req.ProtoMajor))
		log.Printf("ERR: %s: %s [%s]", "Auth", ErrPanic, err)
		return
	}
	// 判断代理的情形
	if req.Method == http.MethodConnect {
		s.onRequestProxy(ctx, res, req)
	} else {
		s.proxyCheatAddress(res, req)
	}
}
