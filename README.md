# Parrot

Parrot is a lightweight, single-threaded TCP echo server written in Go. It demonstrates high-concurrency handling using low-level system calls and I/O multiplexing without relying on Go's standard `net` package or goroutines for connection handling.

## Features

- **Single Threaded**: Handles all connections in the main thread.
- **I/O Multiplexing**: Uses `kqueue` for efficient event-driven I/O.
- **Non-Blocking**: All socket operations are non-blocking.
- **Zero Dependencies**: Uses only the Go standard library (`syscall`).
- **Configurable**: Command-line flag for port selection.
- **Logging**: Extended CLF-style logging for received messages.

## Compatibility

> [!WARNING]
> **macOS & BSD Only**

This project is implemented using **`kqueue`**, which is specific to BSD-derived operating systems. It **will not work** on Linux (which uses `epoll`) or Windows (which uses IOCP).

Supported Platforms:
- macOS (Darwin)
- FreeBSD
- OpenBSD
- NetBSD
- DragonflyBSD

## Usage

### Prerequisites
- Go installed on a supported OS (macOS/BSD).

### Running the Server
```bash
go run main.go -port 1512
```
*Default port is 1512 if not specified.*

### Testing
You can test the server using `nc` (netcat) in multiple terminal windows to verify concurrent handling:

```bash
# Terminal 1
nc localhost 1512
```

```bash
# Terminal 2
nc localhost 1512
```

## Implementation Details
The server bypasses Go's high-level `net` package and interacts directly with the kernel using `syscall`:
1.  **Socket Creation**: `syscall.Socket`
2.  **Non-Blocking Mode**: `syscall.SetNonblock`
3.  **Event Loop**: `syscall.Kqueue` & `syscall.Kevent`
