// Author: Om Patil <patilom001@gmail.com>
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"syscall"
)

func main() {
	// Start pprof server
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	port := flag.Int("port", 1512, "Port to listen on")
	flag.Parse()

	// 1. Create Socket (IPv4, TCP)
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer syscall.Close(serverFD)

	// 2. Set SO_REUSEADDR
	if err := syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatalf("Failed to set SO_REUSEADDR: %v", err)
	}

	// 3. Set Non-Blocking
	if err := syscall.SetNonblock(serverFD, true); err != nil {
		log.Fatalf("Failed to set non-blocking: %v", err)
	}

	// 4. Bind to 0.0.0.0:port
	addr := &syscall.SockaddrInet4{Port: *port}
	copy(addr.Addr[:], []byte{0, 0, 0, 0})
	if err := syscall.Bind(serverFD, addr); err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}

	// 5. Listen
	if err := syscall.Listen(serverFD, 128); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Printf("Server listening on 0.0.0.0:%d (kqueue)\n", *port)

	// 6. Create Kqueue
	kq, err := syscall.Kqueue()
	if err != nil {
		log.Fatalf("Failed to create kqueue: %v", err)
	}
	defer syscall.Close(kq)

	// 7. Register Listen Event
	change := syscall.Kevent_t{
		Ident:  uint64(serverFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}
	if _, err := syscall.Kevent(kq, []syscall.Kevent_t{change}, nil, nil); err != nil {
		log.Fatalf("Failed to register kevent: %v", err)
	}

	// Buffer for reading (reused)
	buf := make([]byte, 1024)

	// Event Loop
	events := make([]syscall.Kevent_t, 20)
	for {
		// Wait for events
		n, err := syscall.Kevent(kq, nil, events, nil)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			log.Fatalf("Kevent wait failed: %v", err)
		}

		for i := 0; i < n; i++ {
			event := events[i]
			fd := int(event.Ident)

			if fd == serverFD {
				// Accept new connection
				connFD, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Printf("Accept failed: %v", err)
					continue
				}

				if err := syscall.SetNonblock(connFD, true); err != nil {
					syscall.Close(connFD)
					continue
				}

				// Register Read Event for new connection
				connChange := syscall.Kevent_t{
					Ident:  uint64(connFD),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
				}
				if _, err := syscall.Kevent(kq, []syscall.Kevent_t{connChange}, nil, nil); err != nil {
					syscall.Close(connFD)
					continue
				}

			} else if event.Filter == syscall.EVFILT_READ {
				// Read from client
				nRead, err := syscall.Read(fd, buf)
				if err != nil || nRead == 0 {
					// EOF or Error
					syscall.Close(fd)
					continue
				}

				// Log message
				// remoteAddr := getPeerName(fd)
				// timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
				// fmt.Printf("%s - - [%s] %q\n", remoteAddr, timestamp, string(buf[:nRead]))

				// Echo back
				_, err = syscall.Write(fd, buf[:nRead])
				if err != nil {
					log.Printf("Write error: %v", err)
					syscall.Close(fd)
				}
			}
		}
	}
}

// sockAddrToString converts a syscall.Sockaddr to a string representation
func sockAddrToString(sa syscall.Sockaddr) string {
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		return fmt.Sprintf("%d.%d.%d.%d:%d", v.Addr[0], v.Addr[1], v.Addr[2], v.Addr[3], v.Port)
	case *syscall.SockaddrInet6:
		// Simplified IPv6 format
		return fmt.Sprintf("[IPv6]:%d", v.Port)
	}
	return "unknown"
}

// getPeerName returns the peer address (remote address) of a file descriptor
func getPeerName(fd int) string {
	sa, err := syscall.Getpeername(fd)
	if err != nil {
		return "-"
	}
	return sockAddrToString(sa)
}
