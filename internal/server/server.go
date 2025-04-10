package server

import (
	"bytes"
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/request"
	"github.com/alexmarian/httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener *net.Listener
	handler  *Handler
	closed   atomic.Bool
}

type Handler func(w io.Writer, req *request.Request) *response.HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error listening on port %d: %w", port, err)
	}
	server := &Server{
		listener: &listener,
		handler:  &handler,
		closed:   atomic.Bool{},
	}
	go server.listen(handler)
	return server, nil
}
func (s *Server) Close() error {
	s.closed.Store(true)
	if err := (*s.listener).Close(); err != nil {
		return fmt.Errorf("error closing server: %w", err)
	}
	return nil
}
func (s *Server) listen(handler Handler) {
	for !s.closed.Load() {
		conn, err := (*s.listener).Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("Accepted connection: %v\n", conn)
		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Error reading request:", err)
		response.HandlerError{
			StatusCode: response.BAD_REQUEST,
			Message:    err.Error(),
		}.Write(conn)
		return
	}
	buf := bytes.NewBuffer([]byte{})
	hErr := (*s.handler)(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}
	b := buf.Bytes()
	response.WriteStatusLine(conn, response.SUCCESS)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
	return
}
