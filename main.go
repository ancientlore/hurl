package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/ancientlore/flagcfg"
	"github.com/ancientlore/kubismus"
	"github.com/facebookgo/flagenv"
)

type hdrMode int

const (
	hdrSet hdrMode = iota
	hdrAdd
)

type hdr struct {
	Key   string
	Value string
	Mode  hdrMode
}

var (
	client       *http.Client
	transport    *http.Transport
	addr         string
	conns        int
	timeout      time.Duration
	method       string
	loop         int
	filesPat     string
	cpuProfile   string
	memProfile   string
	traceFile    string
	cpus         int
	workingDir   string
	discard      bool
	noCompress   bool
	noKeepAlive  bool
	useRequestID string
	headerDelim  string
	headerText   string
	version      bool
	help         bool
	headers      []hdr
)

//go:embed media/*.png
var media embed.FS

func init() {

	// http service/status address
	flag.StringVar(&addr, "addr", ":8080", "HTTP service address for monitoring.")

	// http post settings
	flag.IntVar(&conns, "conns", 2, "Number of concurrent HTTP connections.")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "HTTP timeout.")
	flag.StringVar(&filesPat, "files", "", "Pattern of files to post, like *.xml. Comma-separate for multiple patterns.")
	flag.StringVar(&method, "method", "GET", "HTTP method.")
	flag.StringVar(&useRequestID, "requestid", "", "Name of header to send a random GUID.")
	flag.IntVar(&loop, "loop", 1, "Number of times to loop and repeat.")
	flag.BoolVar(&discard, "discard", false, "Discard received data.")
	flag.BoolVar(&noCompress, "nocompress", false, "Disable HTTP compression.")
	flag.BoolVar(&noKeepAlive, "nokeepalive", false, "Disable HTTP keep-alives.")

	// headers
	flag.StringVar(&headerDelim, "hdrdelim", "|", "Delimiter for HTTP headers specified with -header.")
	flag.StringVar(&headerText, "headers", "", "HTTP headers, delimited by -hdrdelim.")

	// profiling
	flag.StringVar(&cpuProfile, "cpuprofile", "", "Write CPU profile to given file.")
	flag.StringVar(&memProfile, "memprofile", "", "Write memory profile to given file.")
	flag.StringVar(&traceFile, "trace", "", "Write a trace file.")

	// runtime
	flag.IntVar(&cpus, "cpu", 0, "Number of CPUs to use.")
	flag.StringVar(&workingDir, "wd", "", "Set the working directory.")

	// help
	flag.BoolVar(&version, "version", false, "Show version.")
	flag.BoolVar(&help, "help", false, "Show help.")
}

func showHelp() {
	fmt.Println(`
    __    __  ______  __ 
   / /_  / / / / __ \/ / 
  / __ \/ / / / /_/ / /  
 / / / / /_/ / _, _/ /___
/_/ /_/\____/_/ |_/_____/

A tool to fetch over HTTP, slanted towards load generation.

Usage:
  hurl [options] url1 [url2 ... urlN]
  hurl [options] @urlFile

Example:
  hurl -method POST -files "*.xml" -conns 10 http://localhost/svc/foo http://localhost/svc/bar
  hurl -method POST -files "*.xml" -conns 10 @urls.txt
  hurl -method POST -files "*.xml" -conns 10 < urls.txt

Options:`)
	flag.PrintDefaults()
	fmt.Println(`
All of the options can be set via environment variables prefixed with "HURL_" - for instance,
HURL_TIMEOUT can be set to "30s" to increase the default timeout.

Options can also be specified in a TOML configuration file named "hurl.config". The location
of the file can be overridden with the HURL_CONFIG environment variable.`)
}

func showVersion() {
	fmt.Printf("hURL version %s\n", hurlVersion)
}

func parseHeaders() error {
	headers = make([]hdr, 0)
	headerText = strings.TrimSpace(headerText)
	if headerText != "" {
		arr := strings.Split(headerText, headerDelim)
		found := make(map[string]bool)
		for _, h := range arr {
			harr := strings.SplitN(h, ":", 2)
			if len(harr) != 2 {
				return errors.New("Unable to parse header: " + h)
			}
			newHdr := hdr{Key: strings.TrimSpace(harr[0]), Value: strings.TrimSpace(harr[1])}
			_, ok := found[newHdr.Key]
			if !ok {
				found[newHdr.Key] = true
				newHdr.Mode = hdrSet
			} else {
				newHdr.Mode = hdrAdd
			}
			headers = append(headers, newHdr)
		}
	}
	return nil
}

func main() {
	// Parse flags from command-line
	flag.Parse()

	// Parser flags from config
	flagcfg.AddDefaults()
	flagcfg.Parse()

	// Parse flags from environment (using github.com/facebookgo/flagenv)
	flagenv.Prefix = "HURL_"
	flagenv.Parse()

	if help {
		showHelp()
		return
	}

	if version {
		showVersion()
		return
	}

	err := parseHeaders()
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("%#v", headers)

	// setup number of CPUs
	runtime.GOMAXPROCS(cpus)

	// setup cpu profiling if desired
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer func() {
			log.Print("Writing CPU profile to ", cpuProfile)
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	if traceFile != "" {
		f, err := os.Create(traceFile)
		if err != nil {
			log.Fatalf("failed to create trace output file: %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("failed to close trace file: %v", err)
			}
		}()

		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	}

	// create HTTP transport and client
	transport = &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: noKeepAlive, MaxIdleConnsPerHost: conns, DisableCompression: noCompress, ResponseHeaderTimeout: timeout}
	client = &http.Client{Transport: transport, Timeout: timeout}

	// setup Kubismus
	kubismus.Setup("hURL", "/media/logo36.png")
	kubismus.Define("Sent", kubismus.COUNT, "HTTP Posts")
	kubismus.Define("Sent", kubismus.SUM, "Bytes Sent")
	kubismus.Define("Received", kubismus.SUM, "Bytes Received")
	kubismus.Define("Received100", kubismus.COUNT, "1xx Responses")
	kubismus.Define("Received200", kubismus.COUNT, "2xx Responses")
	kubismus.Define("Received300", kubismus.COUNT, "3xx Responses")
	kubismus.Define("Received400", kubismus.COUNT, "4xx Responses")
	kubismus.Define("Received500", kubismus.COUNT, "5xx Responses")
	kubismus.Define("Error", kubismus.COUNT, "Communication Errors")
	kubismus.Define("ResponseTime", kubismus.AVERAGE, "Average Time (s)")
	kubismus.Note("Concurrent Connections", strconv.Itoa(conns))
	kubismus.Note("HTTP Method", method)
	kubismus.Note("Timeout", timeout.String())
	kubismus.Note("Processors", fmt.Sprintf("%d of %d", runtime.GOMAXPROCS(0), runtime.NumCPU()))
	kubismus.Note("Data files", strings.Join(strings.Split(filesPat, ","), "\n"))
	kubismus.Note("URLs", strings.Join(flag.Args(), "\n"))
	kubismus.Note("Discard files", strconv.FormatBool(discard))
	http.Handle("/", http.HandlerFunc(kubismus.ServeHTTP))
	http.Handle("/media/", http.FileServer(http.FS(media)))

	// switch to working dir
	if workingDir != "" {
		err := os.Chdir(workingDir)
		if err != nil {
			log.Fatal(err)
		}
	}
	wd, err := os.Getwd()
	if err == nil {
		kubismus.Note("Working Directory", wd)
	}

	// setup the thread context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// spawn a function that updates the number of goroutines shown in the status page
	go func() {
		done := ctx.Done()
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				kubismus.Note("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
			}
		}
	}()

	// spawn the status web site
	go func() {
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	// handle kill signals
	go func() {
		// Set up channel on which to send signal notifications.
		// We must use a buffered channel or risk missing the signal
		// if we're not ready to receive when the signal is sent.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		// Block until a signal is received.
		s := <-c
		log.Print("Got signal ", s, ", canceling work")
		cancel()
	}()

	// Build pipeline
	var patList []string
	if filesPat != "" {
		patList = strings.Split(filesPat, ",")
	}
	ch1 := loopCount(ctx, loop)
	ch2 := loopUrls(ctx, method, flag.Args(), ch1)
	ch3 := loopFiles(ctx, patList, ch2)
	doHTTP(ctx, conns, ch3)

	// write memory profile if configured
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Print(err)
		} else {
			log.Print("Writing memory profile to ", memProfile)
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}
}
