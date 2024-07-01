package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

type server interface {
	Start(addr string) error
	HandleConns() error
	Stop()
}

type serverImpl struct {
	listener net.Listener
}

type clientHandler interface {
	Handle(net.Conn)
	Broadcast()
}

type client struct {
	name     string
	outgoing chan string
}

type clientHandlerImpl struct {
	incoming chan string
	outgoing map[string]chan string
	joined   chan client
	left     chan string
}

func (s *serverImpl) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = l

	log.Printf("Listening on %q", s.listener.Addr())

	return nil
}

func (s *serverImpl) HandleConns(handler clientHandler) error {
	go handler.Broadcast()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		handler.Handle(conn)
	}
}

func (s *serverImpl) Stop() {
	s.listener.Close()
}

// Handle recieves incoming messages from a client.
func (h *clientHandlerImpl) Handle(conn net.Conn) {
	c := client{
		name:     conn.RemoteAddr().String(),
		outgoing: make(chan string),
	}

	h.joined <- c

	go h.Receive(conn)
	go h.Send(conn, c.outgoing)
}

func (h *clientHandlerImpl) Receive(conn net.Conn) {
	clientName := conn.RemoteAddr()
	fmt.Fprintf(conn, "Hello, %s\n", clientName)

	inputScanner := bufio.NewScanner(conn)
	for inputScanner.Scan() {
		message := fmt.Sprintf("%s: %s\n", clientName, inputScanner.Text())
		h.incoming <- message
	}

	h.left <- clientName.String()
}

func (h *clientHandlerImpl) Send(conn net.Conn, outgoing <-chan string) {
	for message := range outgoing {
		fmt.Fprintf(conn, message)
	}
}

func (h *clientHandlerImpl) Broadcast() {
	for {
		select {
		case message := <-h.incoming:
			for _, outgoing := range h.outgoing {
				outgoing <- message
			}
		case c := <-h.joined:
			log.Printf("%q joined\n", c.name)
			h.outgoing[c.name] = c.outgoing
		case c := <-h.left:
			log.Printf("%q left\n", c)
			close(h.outgoing[c])
			delete(h.outgoing, c)
		}
	}
}

func main() {
	port := flag.Int("port", 8000, "port to run on")
	flag.Parse()

	s := &serverImpl{}

	listenAddr := fmt.Sprintf("localhost:%d", *port)
	err := s.Start(listenAddr)
	if err != nil {
		log.Fatal("Failed to listen on %q: %v", listenAddr, err)
	}
	defer s.Stop()

	handler := &clientHandlerImpl{
		incoming: make(chan string),
		outgoing: map[string]chan string{},
		joined:   make(chan client),
		left:     make(chan string),
	}
	s.HandleConns(handler)
}
