package main

import (
	"flag"
	"log"
	"net"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:1512", "Target address")
	count := flag.Int("c", 1000, "Number of connections")
	flag.Parse()

	conns := make([]net.Conn, *count)
	for i := 0; i < *count; i++ {
		c, err := net.Dial("tcp", *addr)
		if err != nil {
			log.Printf("Dial failed at %d: %v", i, err)
			break
		}
		conns[i] = c
	}
	log.Printf("Established %d connections. Sleeping...", *count)
	for {
		time.Sleep(10 * time.Second)
	}
}
