package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ancientlore/kubismus"
	"github.com/google/uuid"
)

var (
	re *regexp.Regexp
)

func init() {
	re = regexp.MustCompile(`[^\w]+`)
}

func urlToFilename(i *hRequest) string {
	rawurl := strings.TrimPrefix(i.URL, "http://")
	rawurl = strings.TrimPrefix(rawurl, "https://")
	s := re.ReplaceAllString(rawurl, "_")
	s = strings.Trim(s, "_")
	s += "_" + re.ReplaceAllString(i.Filename, "_")
	s = strings.Trim(s, "_")
	return s + "_" + strconv.Itoa(i.LoopNum) + ".out"
}

func doHTTP(ctx context.Context, postThreads int, ch <-chan hRequest) {
	var wg sync.WaitGroup

	// create HTTP posting threads
	wg.Add(postThreads)
	for i := 0; i < postThreads; i++ {
		go posterThread(ctx, ch, &wg)
	}

	// Wait for threads to finish
	wg.Wait()
}

func posterThread(ctx context.Context, ch <-chan hRequest, wg *sync.WaitGroup) {
	done := ctx.Done()
	defer wg.Done()

	for {
		select {
		case i, ok := <-ch:
			if !ok {
				return
			}
			//log.Printf("%#v", i)
			tr := trace.StartRegion(ctx, fmt.Sprintf("#%d %s %s", i.LoopNum, i.Method, i.URL))
			var f io.ReadCloser
			var err error
			if i.Filename != "" {
				f, err = os.Open(i.Filename)
				if err != nil {
					log.Printf("Unable to open %s", i.Filename)
					tr.End()
					continue
				}
			}
			req, err := http.NewRequest(i.Method, i.URL, f)
			if err != nil {
				if f != nil {
					f.Close()
				}
				tr.End()
				log.Fatal(err)
			}
			if i.Filename != "" {
				ct := mime.TypeByExtension(filepath.Ext(i.Filename))
				if ct != "" {
					req.Header.Set("Content-Type", ct)
				}
				req.ContentLength = i.Size
			}
			if useRequestID != "" {
				guid, err := uuid.NewRandom()
				if err == nil {
					req.Header.Set(useRequestID, guid.String())
				}
			}
			for _, h := range headers {
				if h.Mode == hdrSet {
					req.Header.Set(h.Key, h.Value)
				} else {
					req.Header.Add(h.Key, h.Value)
				}
			}
			req.Close = false
			// log.Printf("%#v", req)
			t := time.Now()
			resp, err := client.Do(req)
			if f != nil {
				f.Close()
			}
			if err != nil {
				kubismus.Metric("Error", 1, 0)
				log.Print("HTTP error ", i.URL, ": ", err)
				tr.End()
				continue
			}
			kubismus.Metric("Sent", 1, float64(i.Size))
			name := urlToFilename(&i)
			// log.Print("File would be ", name)
			var outfile io.WriteCloser
			writeTo := ioutil.Discard
			if !discard {
				outfile, err := os.Create(name)
				if err != nil {
					log.Print("Unable to create file ", name)
				} else {
					writeTo = outfile
				}
			}
			if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
				log.Print("Failed to post to ", i.URL, ", status ", resp.Status)
			}
			//if resp.ContentLength > 0 {
			statusRange := resp.StatusCode / 100
			switch statusRange {
			case 1:
				kubismus.Metric("Received100", 1, 0)
			case 2:
				kubismus.Metric("Received200", 1, 0)
			case 3:
				kubismus.Metric("Received300", 1, 0)
			case 4:
				kubismus.Metric("Received400", 1, 0)
			case 5:
				kubismus.Metric("Received500", 1, 0)
			}
			sz, err := io.Copy(writeTo, resp.Body)
			if err == nil {
				kubismus.Metric("Received", 1, float64(sz))
			}
			//}
			resp.Body.Close()
			d := time.Since(t)
			kubismus.Metric("ResponseTime", 1, float64(d.Nanoseconds())/float64(time.Second))
			if outfile != nil {
				outfile.Close()
			}
			tr.End()
		case <-done:
			return
		}
	}
}
