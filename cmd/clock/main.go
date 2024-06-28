package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func handleConn(conn net.Conn) error {
	for {
		t := time.Now()
		if _, err := io.WriteString(conn, t.Format("15:04:05\n")); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func main() {
	port := flag.Int("port", 8000, "port to listen on")

	flag.Parse()

	listenAddr := fmt.Sprintf("localhost:%d", *port)
	log.Printf("listening for connections on %q", listenAddr)

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("failed to setup TCP listener: %v", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
		}

		go func() {
			if err := handleConn(conn); err != nil {
				log.Printf("failed to handle connection: %v", err)
			}
		}()

	}
}
