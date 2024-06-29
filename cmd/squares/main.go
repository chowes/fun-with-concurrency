package main

import (
	"fmt"
)

func iterator(naturals chan<- int, n int) {
	for i := range n {
		naturals <- i + 1
	}
	close(naturals)
}

func squarer(naturals <-chan int, squares chan<- int) {
	for i := range naturals {
		squares <- i * i
	}
	close(squares)
}

func printer(squares <-chan int, done chan<- bool) {
	for i := range squares {
		fmt.Println(i)
	}
	done <- true
}

func main() {
	naturals := make(chan int)
	squares := make(chan int)
	done := make(chan bool)

	go func() {
		iterator(naturals, 1000000)
	}()

	go func() {
		squarer(naturals, squares)
	}()

	go func() {
		printer(squares, done)
	}()

	<-done
}
