package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
)

func loopUrls(ctx context.Context, method string, urls []string, ch <-chan hRequest) <-chan hRequest {
	done := ctx.Done()
	out := make(chan hRequest)
	urlsInFile := func(i hRequest, mth string, fileName string, c <-chan hRequest) {
		f, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" {
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
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	looper := func(mth string, urlList []string, c <-chan hRequest) {
		defer close(out)
		for i := range c {
			for _, url := range urlList {
				// check for file of URLs
				if strings.HasPrefix(url, "@") {
					urlsInFile(i, mth, url[1:], c)
				} else {
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
	}
	go looper(method, urls, ch)
	return out
}
