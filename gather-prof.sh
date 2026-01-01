
rm prof.log 

while [ true ]; do 
    date | tee -a prof.log
    curl http://localhost:6060/debug/pprof/goroutine?debug=1 | tee -a prof.log
    sleep 1s
done
