package counter

import (
	"sync/atomic"
)

type Counter struct {
	threads   int64
	localVals []atomic.Int64
}

func New(threads int64) *Counter {
	return &Counter{
		threads:   threads,
		localVals: make([]atomic.Int64, threads, threads),
	}
}

func (c *Counter) Increment(threadID int64) {
	c.localVals[threadID].Add(1)
}

func (c *Counter) Read() int64 {
	var sum int64 = 0

	for i := range c.threads {
		sum += c.localVals[i].Load()
	}

	return sum
}
