## How to a profile a go profile with pprof library 

# Profiles

The official documentations [heap profiles](https://pkg.go.dev/runtime/pprof#hdr-Heap_profile-Profile)

Go profiles are collection of stack traces taken from a running go program, that we want to analyze. There are different kinds of go profiles, as will be explained later.

## Enable profile collection

You need to add the following code to your executable:

```go

import _ "net/http/pprof"
import "net/http"
import "os"

func main() {
    # need to start this initialization for using prof... 
    if os.getenv("LISTEN_PPROF") != nil {
        go func() {
            http.ListenAndServe("localhost:6060", nil)
        }()
    }
}
```

in more detail:

- Here profiling is enabled, on condition that  `LISTEN_PPROF` environment variable is set 
- it starts a go routine that starts a web server on port 6060 (http.ListenAndServer). A go routine is running in the background.
- note that a nil parameter is passed for the function parameter that that would normally get a function for serving of the http request. This means that the default handler [DefaultServeMux](https://pkg.go.dev/net/http#DefaultServeMux) is called
- important detail: `import _ "net/http/pprof"` includes the `pprof` package - this tells DefaultServeMux how to handle http requests for information.

Lets look at the source code [link](https://cs.opensource.google/go/go/+/refs/tags/go1.25.5:src/net/http/pprof/pprof.go)

- when the package is included, the `init` function it runs. It looks as follows:

```go
func init() {
	prefix := ""
	if godebug.New("httpmuxgo121").Value() != "1" {
		prefix = "GET "
	}
	http.HandleFunc(prefix+"/debug/pprof/", Index)
	http.HandleFunc(prefix+"/debug/pprof/cmdline", Cmdline)
	http.HandleFunc(prefix+"/debug/pprof/profile", Profile)
	http.HandleFunc(prefix+"/debug/pprof/symbol", Symbol)
	http.HandleFunc(prefix+"/debug/pprof/trace", Trace)
}
```

Go profiling is served by `http.HandleFunc(prefix+"/debug/pprof/profile", Profile)`

However the other handlers are interesting too: 
- `/debug/pprof/cmdline`  (access by curl http://localhost:6060/debug/pprof/cmdliine) gets you the command line parameters of the running process, separated by NULL bytes.
- pattern `/debug/pprof/`  (access by curl http://localhost:6060/debug/pprof) gives you an html page that lists all available kinds of profiles, with a description, that sometimes rivals that of the official documentation in clarity!

- each profile is the parameter that comes next in the url. The following profiles return information on the current state of what happens at the time of the the http request:

Url parameter is `debug=1` - this tells it to return text formatted output, which is easier to look at compared to default binary output.

<table>
  <tr>
    <td>
        goroutine
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/goroutine?debug=1'`
    </td>
    <td> 
        The stack trace of all running go routines
    </td>
  </tr>
  <tr>
    <td>
        mutex
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/mutex?debug=1'`
    </td>
    <td> 
        stack trace of go routines that are now holding a mutex / are now waiting on a synchronization primitive
    </td>
  </tr>
</table>

Other profiles return some accumulated information of past events, but you can pass an additional parameter that requests information for events during the past N seconds by passing an additional seconds=N parameter to the URL!

This seconds=N parameter works for: allocs, block, goroutine, heap, mutex, threadcreate


<table>
  <tr>
    <td>
        threadcreate
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/threadcreate?seconds=3&debug=1'`
    </td>
    <td> 
        The stack trace of go routines that created a goroutine/thread during the last three seconds 
    </td>
  </tr>

  <tr>
    <td>
        alloc
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/alloc?seconds=3&debug=1'`
    </td>
    <td> 
        The stack trace of go routines that performed an allocation during the last three seconds (isn't that all of them?)
    </td>
  </tr>

  <tr>
    <td>
        heap
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/alloc?seconds=3&debug=1'`
    </td>
    <td> 
        subset of alloc, only those threads that produced a currently living objects are counted here.
    </td>
  </tr>
  <tr>
    <td>
        block
    </td>
    <td>    
        `curl 'http://localhost:6060/debug/pprof/block?seconds=3&debug=1'`
    </td>
    <td> 
        `block` profile is a superset of `mutex` - here a all goroutines count that have resulted in a blocking - for last three seconds
    </td>
  </tr>
</table>


# Gathering 

The following one line script gathers result into file ```prof.log``` - every second you get a new entry into the file. This script uses the `goroutine` profile - which lists all running go routines.

<code>
<pre>
rm prof.log; while [ true ]; do date | tee -a prof.log; curl http://localhost:6060/debug/pprof/goroutine?debug=1 | tee -a prof.log ; sleep 1s; done
</pre>
</code>

First thing you want to do: see if the number of go routine does not grow without bounds. Leaking go routines is a very bad thing, for performance - and should be fixed as a first priority.

Most simple way as follows:

<code>
<pre>
grep -B 1 'goroutine profile:' prof.log
</pre>
</code>

This gives you something like this:

<code>
<pre>
--
Wed Dec 31 13:49:53 IST 2025
goroutine profile: total 76
--
Wed Dec 31 13:49:54 IST 2025
goroutine profile: total 76
--
Wed Dec 31 13:49:55 IST 2025
goroutine profile: total 76
</pre>
</code>


## Tool for displaying a call graph with number of calls for each frame.

In this repository build the program `parseprof.go`

<pre>
<code>
go build -o parseprof
</code>
</pre>

Run the program to create an html file for the call graph

<pre>
<code>
./parseprof -in prof.log -out prof.html
</code>
</pre>

Display the resulting html file in a web browser


