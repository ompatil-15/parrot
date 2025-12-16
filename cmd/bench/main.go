package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:1512", "Target address")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	totalRequests := flag.Int("n", 10000, "Total number of requests")
	keepAlive := flag.Bool("k", true, "Keep connections alive")
	flag.Parse()

	fmt.Printf("Benchmarking %s with %d connections, %d requests total...\n", *addr, *concurrency, *totalRequests)

	requestsPerWorker := *totalRequests / *concurrency
	var wg sync.WaitGroup
	results := make(chan time.Duration, *totalRequests)

	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(*addr, requestsPerWorker, *keepAlive, results)
		}()
	}

	wg.Wait()
	close(results)
	totalTime := time.Since(start)

	var latencies []time.Duration
	for lat := range results {
		latencies = append(latencies, lat)
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	printStats(totalTime, len(latencies), latencies)
}

func worker(addr string, attempts int, keepAlive bool, results chan<- time.Duration) {
	msg := []byte("Hello Parrot")
	buf := make([]byte, 1024)

	var conn net.Conn
	var err error

	if keepAlive {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			log.Printf("Dial failed: %v", err)
			return
		}
		defer conn.Close()
	}

	for i := 0; i < attempts; i++ {
		if !keepAlive {
			conn, err = net.Dial("tcp", addr)
			if err != nil {
				log.Printf("Dial failed: %v", err)
				continue
			}
		}

		start := time.Now()
		if _, err := conn.Write(msg); err != nil {
			log.Printf("Write failed: %v", err)
			if !keepAlive {
				conn.Close()
			}
			return
		}

		if _, err := conn.Read(buf); err != nil {
			log.Printf("Read failed: %v", err)
			if !keepAlive {
				conn.Close()
			}
			return
		}

		results <- time.Since(start)

		if !keepAlive {
			conn.Close()
		}
	}
}

func printStats(totalTime time.Duration, count int, latencies []time.Duration) {
	if count == 0 {
		fmt.Println("No successful requests.")
		return
	}
	rps := float64(count) / totalTime.Seconds()
	p50 := latencies[count*50/100]
	p90 := latencies[count*90/100]
	p99 := latencies[count*99/100]

	fmt.Printf("\n--- Results ---\n")
	fmt.Printf("Total Requests: %d\n", count)
	fmt.Printf("Total Time:     %v\n", totalTime)
	fmt.Printf("RPS:            %.2f\n", rps)
	fmt.Printf("Latency P50:    %v\n", p50)
	fmt.Printf("Latency P90:    %v\n", p90)
	fmt.Printf("Latency P99:    %v\n", p99)
}
