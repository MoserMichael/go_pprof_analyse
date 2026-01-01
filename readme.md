## How to a profile a go profile with pprof library 

The go profile can respond to http requests by sending statistics, while requested via http. That's why you need to add the following to your executable: (starts to listen for external requests for pprof statistics, if environment ```LISTEN_PPROF``` is defined)

<code>
<pre>
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
</pre>
</code>


# Gathering 

The folling one line script gathers result into file ```prof.log``` - every second you get a new entry into the file.

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


