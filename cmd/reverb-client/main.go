package main

import (
	"io"
	"log"
	"net"
	"os"
)

type echoClient interface {
	Echo() error
}

type echoClientImpl struct {
	conn net.Conn
}

func newEchoClient(addr string) (*echoClientImpl, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &echoClientImpl{
		conn: conn,
	}, nil
}

func (c *echoClientImpl) Echo() error {
	return echo(os.Stdin, c.conn)
}

func echo(r io.Reader, w io.Writer) error {
	for {
		_, err := io.Copy(w, r)
		if err != nil {
			return err
		}
	}
}

func main() {
	c, err := newEchoClient("localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	c.Echo()
}
