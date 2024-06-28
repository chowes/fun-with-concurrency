package main

import (
	"fmt"
	"time"
)

func spinner(delay time.Duration) {
	for {
		for _, c := range `-\|/` {
			fmt.Printf("\r%c", c)
			time.Sleep(delay)
		}
	}
}

func fib(n int) int {
	if n < 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	go spinner(100 * time.Millisecond)
	res := fib(45)
	fmt.Printf("fib(45)=%d\n", res)
}
