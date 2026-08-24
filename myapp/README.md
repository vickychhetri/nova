# Nova IP & Port Scanner — Desktop GUI Tool

A high-performance, concurrent network discovery and port probing desktop application built with the **Nova Go Native Desktop Framework** (`github.com/vickychhetri/nova`).

---

## ⚡ Features

1. **Subnet & CIDR Range Discovery**:
   - Auto-detects your local active network adapter and subnet (e.g. `192.168.1.0/24`).
   - Supports custom CIDR ranges (e.g. `10.0.0.0/24`, `172.16.0.0/24`, `127.0.0.1/32`).
2. **Multi-Port Probing**:
   - Scans common service ports (`22, 80, 443, 8080, 3306, 5432, 6379, 8000, 9000`).
   - Configurable per-probe TCP timeouts (`50ms` to `1000ms`).
3. **High-Throughput Goroutine Worker Pool**:
   - Parallel scan pipeline executing 10–200 concurrent probes simultaneously.
4. **Live Reactive Telemetry**:
   - Real-time animated progress bar ($0\% \rightarrow 100\%$).
   - Live counters for **Discovered Hosts**, **Scanned IPs**, and **Open Ports Found**.
5. **Reverse DNS & Latency Calculation**:
   - Measures exact TCP round-trip latency (e.g. `1.4 ms`).
   - Performs reverse DNS resolution (`net.LookupAddr`) to display hostnames.
6. **Virtualized High-Performance Results Table**:
   - Smoothly displays hundreds of discovered network hosts without UI lag.

---

## 🚀 How to Run & Build

### Development Mode
```bash
cd myapp
./run.sh
```
*or:*
```bash
go run main.go
```

### Build Release Desktop Binary
```bash
cd myapp
./build.sh
```
The compiled standalone executable binary is generated at:
```bash
./bin/ip_scanner
```

---

## 📂 Project Structure

```
myapp/
├── go.mod           # Module configuration
├── main.go          # Complete IP & Port Scanner application
├── build.sh         # Release desktop binary compiler
├── run.sh           # Development launcher
└── README.md        # Documentation
```
