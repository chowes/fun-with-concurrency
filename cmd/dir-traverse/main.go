package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

func monitorForCancel(done chan<- bool) {
	go func() {
		_, err := os.Stdin.Read(make([]byte, 1))
		if err != nil {
			log.Println("failed to monitor stdin for cancellation signal")
		}
		close(done)
	}()
}

func cancelled(done <-chan bool) bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

func startWorker(dirPath string, wg *sync.WaitGroup, fileSizes chan<- int64, errchan chan<- error, tokens chan bool, done <-chan bool) {
	if cancelled(done) {
		return
	}

	wg.Add(1)
	go func() {
		tokens <- true
		defer wg.Done()
		traverse(dirPath, wg, fileSizes, errchan, tokens, done)
		<-tokens
	}()
}

func traverse(dirPath string, wg *sync.WaitGroup, fileSizes chan<- int64, errchan chan<- error, tokens chan bool, done <-chan bool) {
	dirents, err := os.ReadDir(dirPath)
	if err != nil {
		errchan <- fmt.Errorf("failed to list directory %q: %v", dirPath, err)
	}

	for _, dirent := range dirents {
		if dirent.IsDir() {
			startWorker(path.Join(dirPath, dirent.Name()), wg, fileSizes, errchan, tokens, done)
		} else {
			fullPath := path.Join(dirPath, dirent.Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				errchan <- err
				return
			}
			fileSizes <- info.Size()
		}
	}

	return
}

func main() {
	maxThreads := flag.Int("threads", 1, "max number of threads")
	pathRoots := flag.String("path", "", "comma separated list of paths to traverse")
	flag.Parse()

	if *pathRoots == "" {
		log.Fatal("you must specify a path")
	}
	if *maxThreads < 1 {
		log.Fatal("you must specify at least one worker thread")
	}

	var wg sync.WaitGroup
	tokens := make(chan bool, *maxThreads)
	done := make(chan bool)
	fileSizes := make(chan int64, *maxThreads)
	errchan := make(chan error)

	monitorForCancel(done)

	go func() {
		defer close(fileSizes)
		defer close(errchan)

		for _, pathRoot := range strings.Split(*pathRoots, ",") {
			startWorker(pathRoot, &wg, fileSizes, errchan, tokens, done)
		}

		wg.Wait()
	}()

	var totalSize int64
loop:
	for {
		select {
		case _, ok := <-done:
			log.Println("Operation cancelled")
			if !ok {
				done = nil
			}
		case sz, ok := <-fileSizes:
			if !ok {
				fileSizes = nil
				continue
			}
			totalSize += sz
		case err, ok := <-errchan:
			if !ok {
				break loop
			}
			log.Fatalf("failed to traverse %q: %v", *pathRoots, err)
		}
	}

	fmt.Printf("total directory tree size: %d\n", totalSize)
}
