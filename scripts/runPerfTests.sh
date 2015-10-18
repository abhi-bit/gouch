#!/bin/bash

THREADS=1000
CONNECTIONS=1000
TIMEOUT=10
DURATION=60
HOST=$1

#capture throughput for Go prototype
for limit in 1 10 50 100 200 500 1000 2000 3000 5000 10000 20000 30000 50000 70000 100000
do
    wrk -c $CONNECTIONS -t $THREADS -d $DURATION --timeout $TIMEOUT http://$HOST:9093/query?limit=$limit > $limit-go.log
done


#capture throughput for Erlang standalone
for limit in 1 10 50 100 200 500 1000 2000 3000 5000 10000 20000 30000 50000 70000 100000
do
    wrk -c $CONNECTIONS -t $THREADS -d $DURATION --timeout $TIMEOUT http://$HOST:15984/new?limit=$limit > $limit-erl.log
done


