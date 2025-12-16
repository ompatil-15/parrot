package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 1513, "Port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Printf("Standard server listening on %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept failed: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Using io.Copy to echo data back efficiently
	_, err := io.Copy(conn, conn)
	if err != nil {
		// Log only if it's not a normal EOF (optional)
		// log.Printf("Connection error: %v", err)
	}
}
