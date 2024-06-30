package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func receiveCancel(abort chan<- bool) {
	_, err := os.Stdin.Read(make([]byte, 1))
	if err != nil {
		log.Printf("failed to read cancellation input from stdin: %v", err)
	}
	abort <- true
}

func cancelLaunch() {
	fmt.Println("Launch aborted!")
}

func launch() {
	fmt.Println("Liftoff!")
}

func launchSequence() {
	fmt.Println("Commencing countdown...")

	abort := make(chan bool)
	go receiveCancel(abort)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for countdown := 10; countdown > 0; countdown-- {
		fmt.Println(countdown)

		select {
		case <-ticker.C:
			// Do nothing
		case <-abort:
			cancelLaunch()
			return
		}
	}

	launch()
}

func main() {
	launchSequence()
}
