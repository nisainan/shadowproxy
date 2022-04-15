package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"playproxy/util"
)

func (s *Server) onRequestProxy(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	if req.ProtoMajor == 2 {
		if len(req.URL.Scheme) > 0 || len(req.URL.Path) > 0 {
			err := errors.New("CONNECT request has :scheme or/and :path pseudo-header fields")
			log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, err)
			return
		}
	}
	hostPort := req.URL.Host
	if hostPort == "" {
		hostPort = req.Host
	}
	targetConn, dialErr := s.dialContext(ctx, "tcp", hostPort)
	if dialErr != nil {
		//log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, dialErr.Error())
		return
	}
	if targetConn == nil {
		// safest to check both error and targetConn afterwards, in case fp.dial (potentially unstable
		// from x/net/proxy) misbehaves and returns both nil or both non-nil
		err := errors.New("hostname " + req.URL.Hostname() + " is not allowed")
		log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, err)
		return
	}
	defer targetConn.Close()
	switch req.ProtoMajor {
	case 1: // http1: hijack the whole flow
		if _, err := s.serveHijack(res, targetConn); err != nil {
			log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, err)
			return
		}
	case 2: // http2: keep reading from "request" and writing into same response
		defer req.Body.Close()
		wFlusher, ok := res.(http.Flusher)
		if !ok {
			err := errors.New("ResponseWriter doesn't implement Flusher()")
			log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, err)
			return
		}
		res.WriteHeader(http.StatusOK)
		wFlusher.Flush()
		if err := s.dualStream(targetConn, req.Body, res); err != nil {
			//log.Printf("ERR: %s: %s [%s]", "onRequestProxy", ErrPanic, err)
			return
		}
	default:
		panic("There was a check for http version, yet it's incorrect")
	}
	return
}

// Hijacks the connection from ResponseWriter, writes the response and proxies data between targetConn
// and hijacked connection.

func (s *Server) serveHijack(w http.ResponseWriter, targetConn net.Conn) (int, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Hijacker")
	}
	clientConn, bufReader, err := hijacker.Hijack()
	if err != nil {
		return http.StatusInternalServerError, errors.New("failed to hijack: " + err.Error())
	}
	defer clientConn.Close()
	// bufReader may contain unprocessed buffered data from the client.
	if bufReader != nil {
		// snippet borrowed from `proxy` plugin
		if n := bufReader.Reader.Buffered(); n > 0 {
			rbuf, err := bufReader.Reader.Peek(n)
			if err != nil {
				return http.StatusBadGateway, err
			}
			targetConn.Write(rbuf)
		}
	}
	// Since we hijacked the connection, we lost the ability to write and flush headers via w.
	// Let's handcraft the response and send it manually.
	res := &http.Response{StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	err = res.Write(clientConn)
	if err != nil {
		return http.StatusInternalServerError, errors.New("failed to send response to client: " + err.Error())
	}

	return 0, s.dualStream(targetConn, clientConn, clientConn)
}

// Copies data target->clientReader and clientWriter->target, and flushes as needed
// Returns when clientWriter-> target stream is done.
// Server should finish writing target -> clientReader.
func (s *Server) dualStream(target net.Conn, clientReader io.ReadCloser, clientWriter io.Writer) error {
	stream := func(w io.Writer, r io.Reader) error {
		// copy bytes from r to w
		buf := s.bufferPool.Get().([]byte)
		buf = buf[0:cap(buf)]
		_, _err := flushingIoCopy(w, r, buf)
		if closeWriter, ok := w.(interface {
			CloseWrite() error
		}); ok {
			closeWriter.CloseWrite()
		}
		return _err
	}

	go stream(target, clientReader)
	return stream(clientWriter, target)
}

// flushingIoCopy is analogous to buffering io.Copy(), but also attempts to flush on each iteration.
// If dst does not implement http.Flusher(e.g. net.TCPConn), it will do a simple io.CopyBuffer().
// Reasoning: http2ResponseWriter will not flush on its own, so we have to do it manually.
func flushingIoCopy(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	flusher, ok := dst.(http.Flusher)
	if !ok {
		return io.CopyBuffer(dst, src, buf)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			flusher.Flush()
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func (s *Server) proxyCheatAddress(res http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	upsConn, err := s.dialContext(ctx, "tcp", s.confer.CheatHost)
	if err != nil {
		log.Printf("failed to dial upstream : %s\n", err)
		return
	}
	s.forwardResponse(upsConn, res, req)
}

func (s *Server) forwardResponse(conn net.Conn, res http.ResponseWriter, req *http.Request) {
	err := req.Write(conn)
	if err != nil {
		log.Printf("failed to write http request : %s\n", err)
		return
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		log.Printf("failed to read http response : %s\n", err)
		return
	}
	req.Body.Close()
	if response != nil {
		defer response.Body.Close()
	}

	for header, values := range response.Header {
		for _, val := range values {
			res.Header().Add(header, val)
		}
	}
	util.RemoveHopByHop(res.Header())
	res.WriteHeader(response.StatusCode)
	buf := s.bufferPool.Get().([]byte)
	buf = buf[0:cap(buf)]
	io.CopyBuffer(res, response.Body, buf)
	return
}
