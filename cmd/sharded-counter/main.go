package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/chowes/fun-with-concurrency/pkg/counter"
)

func writer(c *counter.Counter, threadID, iterations int64, barrier *sync.RWMutex) {
	barrier.RLock()
	defer barrier.RUnlock()

	for _ = range iterations {
		c.Increment(threadID)
	}
}

func reader(c *counter.Counter, interval time.Duration, maxValue int64, barrier *sync.RWMutex) {
	barrier.RLock()
	defer barrier.RUnlock()

	for {
		v := c.Read()
		fmt.Printf("current counter value is: %d\n", v)
		if v == maxValue {
			return
		}

		time.Sleep(interval)
	}
}

func main() {
	readers := flag.Int64("readers", 0, "number of reader threads")
	writers := flag.Int64("writers", 0, "number of writer threads")
	iterations := flag.Int64("iterations", 0, "number of times each writer should increment the counter")
	readInterval := flag.Duration("interval", 100*time.Millisecond, "the interval each reader should wait between reads")

	flag.Parse()

	c := counter.New(*writers)
	var maxValue int64 = *iterations * *writers
	var wg sync.WaitGroup

	// We use this r/w lock to synchronize thread start times.
	var barrier sync.RWMutex

	barrier.Lock()
	for i := range *writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer(c, i, *iterations, &barrier)
		}()
	}

	for _ = range *readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader(c, *readInterval, maxValue, &barrier)
		}()
	}

	barrier.Unlock()
	startTime := time.Now()
	wg.Wait()

	totalTime := time.Since(startTime)
	fmt.Printf("performed %d increments in %d threads in %s (%.2f ops/second)\n", maxValue, *writers, totalTime.String(), float64(maxValue)/totalTime.Seconds())
}
