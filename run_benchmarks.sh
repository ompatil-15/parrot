#!/bin/bash

# Build Go binaries
go build -o bin/parrot main.go
go build -o bin/standard_server cmd/standard_server/main.go
# No build needed for node

# Increase file descriptor limit for C10K
ulimit -n 20000

CONC_LEVELS=(50 1000 10000)
REQUESTS=5000000

run_test() {
    NAME=$1
    CMD=$2
    PORT=$3
    CONC=$4
    
    echo "------------------------------------------------"
    echo "Running $NAME at $CONC concurrency..."
    
    # Start server
    $CMD &
    PID=$!
    sleep 2

    # Capture pprof for Go servers
    if [[ "$NAME" == *"Go"* ]]; then
        SANITIZED_NAME=$(echo $NAME | tr ' /' '__' | tr -cd 'a-zA-Z0-9_')
        echo "Capturing CPU profile for 10 seconds..."
        curl -s -o ${SANITIZED_NAME}_cpu.prof "http://localhost:6060/debug/pprof/profile?seconds=10" &
        PROFILE_PID=$!
    fi
    
    # Run benchmark
    echo "Performance:"
    go run cmd/bench/main.go -addr localhost:$PORT -c $CONC -n $REQUESTS
    
    # Measure Memory
    MEM=$(ps -o rss= -p $PID | awk '{$1=$1/1024; print $1 " MB"}')
    echo "Memory Usage: $MEM"

    # Wait for profiler if running
    if [[ "$NAME" == *"Go"* ]]; then
        wait $PROFILE_PID
        echo "Profile captured to ${SANITIZED_NAME}_cpu.prof"
        echo "Top CPU consumers:"
        go tool pprof -top ${SANITIZED_NAME}_cpu.prof | head -n 15
    fi
    
    # Kill server
    kill $PID
    sleep 1
}

for C in "${CONC_LEVELS[@]}"; do
    echo "================================================"
    echo "BENCHMARK SUITE: Concurrency $C"
    echo "================================================"
    
    run_test "Parrot (Go/kqueue)" "./bin/parrot" 1512 $C
    run_test "Standard (Go/net)" "./bin/standard_server" 1513 $C
    # run_test "Node.js (v22)" "node cmd/node_server/index.js" 1514 $C
done
