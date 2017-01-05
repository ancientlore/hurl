package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
)

func loopFiles(ctx context.Context, filePatterns []string, ch <-chan hRequest) <-chan hRequest {
	if filePatterns == nil || len(filePatterns) == 0 {
		return ch
	}
	done := ctx.Done()
	out := make(chan hRequest)
	looper := func(patList []string, c <-chan hRequest) {
		defer close(out)
		for i := range c {

			for _, pattern := range patList {
				files, err := filepath.Glob(pattern)
				if err != nil {
					log.Fatal(err)
				}
				for _, f := range files {
					fi, err := os.Stat(f)
					if err != nil {
						log.Fatal(err)
					}
					if !fi.IsDir() {
						/*
							fullpath, err := filepath.Abs(f)
							if err != nil {
								log.Print("Warning, can't get full path of ", f, ": ", err)
								fullpath = f
							}
						*/
						// log.Print("Found ", fullpath)
						data := i
						data.Filename = f
						//data.Filename = fullpath
						data.Size = fi.Size()
						select {
						case out <- data:
						case <-done:
							return
						}

					}
				}
			}
		}
	}
	go looper(filePatterns, ch)
	return out
}
