package main

import (
	"golang.org/x/net/context"
)

func loopUrls(ctx context.Context, method string, urls []string, ch <-chan L) <-chan L {
	done := ctx.Done()
	out := make(chan L)
	looper := func(mth string, urlList []string, c <-chan L) {
		defer close(out)
		for i := range ch {
			for _, url := range urlList {
				data := i
				data.URL = url
				data.Method = mth
				select {
				case out <- data:
				case <-done:
					return
				}
			}
		}
	}
	go looper(method, urls, ch)
	return out
}
