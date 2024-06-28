package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

var (
	tzServerMap = map[string]string{
		"US/Eastern":    "localhost:8000",
		"Asia/Tokyo":    "localhost:8001",
		"Europe/London": "localhost:8002",
	}
)

type timeClient interface {
	MonitorTime() error
	Close()
}

type timeClientImpl struct {
	tz   string
	conn net.Conn
}

func newTimeClient(timeZone string) (*timeClientImpl, error) {
	serverAddr, ok := tzServerMap[timeZone]
	if !ok {
		return nil, fmt.Errorf("%q is not a valid time zone", timeZone)
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	client := &timeClientImpl{
		tz:   timeZone,
		conn: conn,
	}

	return client, nil
}

func (c *timeClientImpl) MonitorTime() error {
	buf := make([]byte, 256)
	for {
		if _, err := c.conn.Read(buf); err != nil {
			return fmt.Errorf("failed to read time: %v", err)
		}

		fmt.Printf("%s: %s", c.tz, string(buf))
	}

	return nil
}

func (c *timeClientImpl) Close() {
	c.conn.Close()
}

func main() {
	var wg sync.WaitGroup

	for tz, _ := range tzServerMap {
		c, err := newTimeClient(tz)
		if err != nil {
			log.Fatalf("failed to create new time client: %v", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.MonitorTime(); err != nil {
				log.Printf("failed to monitor time for timezone %q: %v\n", tz, err)
			}
			defer c.Close()
		}()
	}

	wg.Wait()
}
