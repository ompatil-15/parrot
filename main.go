// Author: Om Patil <patilom001@gmail.com>
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	port := flag.Int("port", 1512, "Port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	// Listen on TCP port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()
	fmt.Printf("Server listening on %s\n", addr)

	for {
		// Accept new connection
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle connection (blocking)
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			break
		}

		timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
		fmt.Printf("%s - - [%s] %q\n", conn.RemoteAddr(), timestamp, string(buf[:n]))

		// Echo back
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}
