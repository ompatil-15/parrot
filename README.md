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

## Benchmarks

We compared Parrot against a standard Go TCP server (`net` package) using a custom benchmarking tool. Tests were run on a MacBook Pro (M-Series).

### Throughput & Efficiency
| Metric | Parot (Single Threaded) | Standard Go (Multi-Threaded) | Difference |
| :--- | :--- | :--- | :--- |
| **Throughput** | **~246,000 RPS** | ~113,000 RPS | **2.2x Faster** |
| **CPU Usage** | **~0.6 Cores** | ~3.5 Cores | **~13x More Efficient** |
| **Memory (C10K)** | **22 MB** | 223 MB | **10x Lower RAM** |

### The C10K Test (10,000 Concurrent Connections)
Parrot demonstrates superior stability under massive concurrency loads.

*   **Parrot**:
    *   Maintained **~219k RPS**.
    *   P99 Latency: **45ms**.
    *   Memory stayed constant at **22 MB**.
*   **Standard Go**:
    *   Throughput dropped to **~109k RPS**.
    *   P99 Latency exploded to **571ms**.
    *   Memory ballooned to **223 MB** (due to goroutine stack overhead).

*> **Key Takeaway**: For I/O-bound workloads, a single-threaded event loop can remarkably outperform multi-threaded architectures in both speed and resource efficiency.*
