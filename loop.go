package main

import (
	"fmt"
	"github.com/ancientlore/kubismus"
	"golang.org/x/net/context"
)

type L struct {
	LoopNum  int
	URL      string
	Method   string
	Filename string
	Size     int64
}

func loopCount(ctx context.Context, count int) <-chan L {
	done := ctx.Done()
	ch := make(chan L)
	looper := func(loopCount int) {
		defer close(ch)
		for i := 1; i <= loopCount; i++ {
			select {
			case <-done:
				return
			case ch <- L{LoopNum: i}:
				kubismus.Note("Loops", fmt.Sprintf("%d of %d", i, loopCount))
			}
		}
	}
	go looper(count)
	return ch
}
