![logo](media/logo72.png)

hURL: A tool to fetch over HTTP, slanted towards load generation.

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
	  -addr=":8080": HTTP service address for monitoring.
	  -conns=10: Number of concurrent HTTP connections.
	  -cpu=2: Number of CPUs to use.
	  -cpuprofile="": Write CPU profile to given file.
	  -discard=false: Discard received data.
	  -files="": Pattern of files to post, like *.xml. Comma-separate for multiple patterns.
	  -hdrdelim="|": Delimiter for HTTP headers specified with -header.
	  -headers="": HTTP headers, delimited by -hdrdelim.
	  -help=false: Show help.
	  -loop=1: Number of times to loop and repeat.
	  -memprofile="": Write memory profile to given file.
	  -method="GET": HTTP method.
	  -nocompress=false: Disable HTTP compression.
	  -nokeepalive=false: Disable HTTP keep-alives.
	  -norequestid=false: Don't send X-RequestID header.
	  -timeout=10s: HTTP timeout.
	  -version=false: Show version.
	  -wd="": Set the working directory.

	All of the options can be set via environment variables prefixed with "HURL_" - for instance,
	HURL_TIMEOUT can be set to "30s" to increase the default timeout.
