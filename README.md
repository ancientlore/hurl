![logo](media/logo72.png)

hURL: A tool to fetch over HTTP, slanted towards load generation.

[![Build Status](https://travis-ci.org/ancientlore/hurl.svg?branch=master)](https://travis-ci.org/ancientlore/hurl)
[![GoDoc](https://godoc.org/github.com/ancientlore/hurl?status.svg)](https://godoc.org/github.com/ancientlore/hurl)

	    __    __  ______  __ 
	   / /_  / / / / __ \/ / 
	  / __ \/ / / / /_/ / /  
	 / / / / /_/ / _, _/ /___
	/_/ /_/\____/_/ |_/_____/

	A tool to fetch over HTTP, slanted towards load generation.

	Usage:
	  hurl [options] url1 [url2 ... urlN]

	Example:
	  hurl -method POST -files "*.xml" -conns 10 http://localhost/svc/foo http://localhost/svc/bar

	Options:
	  -addr string
	    	HTTP service address for monitoring. (default ":8080")
	  -conns int
	    	Number of concurrent HTTP connections. (default 2)
	  -cpu int
	    	Number of CPUs to use.
	  -cpuprofile string
	    	Write CPU profile to given file.
	  -discard
	    	Discard received data.
	  -files string
	    	Pattern of files to post, like *.xml. Comma-separate for multiple patterns.
	  -hdrdelim string
	    	Delimiter for HTTP headers specified with -header. (default "|")
	  -headers string
	    	HTTP headers, delimited by -hdrdelim.
	  -help
	    	Show help.
	  -loop int
	    	Number of times to loop and repeat. (default 1)
	  -memprofile string
	    	Write memory profile to given file.
	  -method string
	    	HTTP method. (default "GET")
	  -nocompress
	    	Disable HTTP compression.
	  -nokeepalive
	    	Disable HTTP keep-alives.
	  -requestid
	    	Send X-RequestID header.
	  -timeout duration
	    	HTTP timeout. (default 10s)
	  -version
	    	Show version.
	  -wd string
	    	Set the working directory.

	All of the options can be set via environment variables prefixed with "HURL_" - for instance,
	HURL_TIMEOUT can be set to "30s" to increase the default timeout.

	Options can also be specified in a TOML configuration file named "hurl.config". The location
	of the file can be overridden with the HURL_CONFIG environment variable.
