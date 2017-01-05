package main

import (
	"context"
	"fmt"

	"github.com/ancientlore/kubismus"
)

type hRequest struct {
	LoopNum  int
	URL      string
	Method   string
	Filename string
	Size     int64
}

func loopCount(ctx context.Context, count int) <-chan hRequest {
	done := ctx.Done()
	ch := make(chan hRequest)
	looper := func(loopCount int) {
		defer close(ch)
		for i := 1; i <= loopCount; i++ {
			select {
			case <-done:
				return
			case ch <- hRequest{LoopNum: i}:
				kubismus.Note("Loops", fmt.Sprintf("%d of %d", i, loopCount))
			}
		}
	}
	go looper(count)
	return ch
}
