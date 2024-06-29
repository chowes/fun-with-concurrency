package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
)

type result struct {
	size int
	err  error
}

// doesn't really do anything, just a mock
func makeThumbnail(path string, results chan<- result) error {
	fmt.Println(path)

	results <- result{
		size: 42,
		err:  fmt.Errorf("foo bar"),
	}

	return nil
}

func makeThumbnails(path string) error {
	results := make(chan result)

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			if err := makeThumbnail(filePath, results); err != nil {
				log.Println(err)
			}
		}(entry.Name())
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		log.Printf("size: %d, err: %v", r.size, r.err)
	}

	return firstErr
}

func main() {
	path := flag.String("path", "", "path to the image directory")
	flag.Parse()

	err := makeThumbnails(*path)
	if err != nil {
		log.Fatal(err)
	}
}
