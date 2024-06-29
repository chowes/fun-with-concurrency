package main

import (
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type echoServer interface {
	Serve()
	Echo(io.Reader, io.Writer) error
}

type echoServerImpl struct {
	delay    time.Duration
	listener net.Listener
}

func newEchoServer(addr string, delay time.Duration) (*echoServerImpl, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &echoServerImpl{
		delay:    delay,
		listener: l,
	}, nil
}

func (s *echoServerImpl) Serve() {
	var wg sync.WaitGroup
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Echo(conn, os.Stdout); err != nil {
				log.Print(err)
			}
		}()
	}
	wg.Wait()
}

func (s *echoServerImpl) sendEcho(w io.Writer, shout string) error {
	if _, err := w.Write([]byte(shout)); err != nil {
		return err
	}

	return nil
}

func (s *echoServerImpl) Echo(r io.Reader, w io.Writer) error {
	for {
		buf := make([]byte, 256)
		if _, err := r.Read(buf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		input := string(buf)

		go func() {
			if err := s.sendEcho(w, strings.ToUpper(input)); err != nil {
				log.Println(err)
			}
			time.Sleep(s.delay)
			if err := s.sendEcho(w, input); err != nil {
				log.Println(err)
			}
			time.Sleep(s.delay)
			if err := s.sendEcho(w, strings.ToLower(input)); err != nil {
				log.Println(err)
			}
		}()
	}

	return nil
}

func main() {
	s, err := newEchoServer("localhost:8000", 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	s.Serve()
}
