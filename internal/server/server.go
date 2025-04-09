package server

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error listening on port %d: %w", port, err)
	}
	server := &Server{
		listener: listener,
		closed:   atomic.Bool{},
	}
	go server.listen()
	return server, nil
}
func (s *Server) Close() error {
	s.closed.Store(true)
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("error closing server: %w", err)
	}
	return nil
}
func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("Accepted connection: %v\n", conn)
		s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	err := response.WriteStatusLine(conn, response.SUCCESS)
	if err != nil {
		log.Println("Error writing status line:", err)
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		log.Println("Error writing headers:", err)
		return
	}
	defer conn.Close()
}
