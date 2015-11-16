package main

import (
	"bufio"
	"golang.org/x/net/context"
	"log"
	"os"
	"strings"
)

func loopUrls(ctx context.Context, method string, urls []string, ch <-chan L) <-chan L {
	done := ctx.Done()
	out := make(chan L)
	urlsInFile := func(i L, mth string, fileName string, c <-chan L) {
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
	looper := func(mth string, urlList []string, c <-chan L) {
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
