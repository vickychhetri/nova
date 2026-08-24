package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/forms"
)

// HostRecord stores detailed network device telemetry.
type HostRecord struct {
	IP        string
	Hostname  string
	Latency   string
	Vendor    string
	Status    string
	OpenPorts string
}

func main() {
	// 1. Initialize Nova Desktop Application
	app := nova.New()

	win := app.Window(
		nova.Title("Nova Network Analyzer & IP Scanner — Enterprise Suite"),
		nova.Size(1260, 840),
		nova.Theme(theme.Dark()),
	)

	// 2. Reactive UI State Signals
	targetSubnet := state.String("192.168.1.0/24")
	targetPorts := state.String("22, 80, 443, 8080, 3306, 5432, 6379, 8000, 9000, 27017")
	timeoutMs := state.Float(200)
	concurrencyLimit := state.Float(60)

	isScanning := state.Bool(false)
	scanProgress := state.Float(0.0)
	hostsDiscovered := state.New([]HostRecord{
		{IP: "192.168.1.1", Hostname: "gateway.router.local", Latency: "1.1 ms", Vendor: "Cisco Systems", Status: "Online", OpenPorts: "80, 443, 53, 22"},
		{IP: "192.168.1.10", Hostname: "fedora-workstation.lan", Latency: "0.4 ms", Vendor: "Dell Inc", Status: "Online", OpenPorts: "22, 8080, 5432"},
		{IP: "192.168.1.45", Hostname: "nas-storage.internal", Latency: "1.8 ms", Vendor: "Synology Inc", Status: "Online", OpenPorts: "80, 443, 5000, 2049"},
		{IP: "192.168.1.102", Hostname: "dev-server-01.corp", Latency: "0.8 ms", Vendor: "Supermicro", Status: "Online", OpenPorts: "22, 80, 443, 3306, 6379"},
	})
	scannedIPCount := state.Int(254)
	totalIPCount := state.Int(254)
	openPortsTotal := state.Int(14)
	statusText := state.String("Ready — 4 active devices in local cache.")
	selectedRowIndex := state.Int(0)

	var cancelScan context.CancelFunc
	var scanMutex sync.Mutex

	// Helper to auto-detect local network IP
	detectLocalSubnet := func() string {
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip := ipnet.IP.To4()
						return fmt.Sprintf("%d.%d.%d.0/24", ip[0], ip[1], ip[2])
					}
				}
			}
		}
		return "192.168.1.0/24"
	}

	targetSubnet.Set(detectLocalSubnet())

	// 3. Scan Worker Engine
	startScan := func(cidr string, portsStr string, timeout time.Duration, concurrency int) {
		scanMutex.Lock()
		if isScanning.Get() {
			scanMutex.Unlock()
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancelScan = cancel
		isScanning.Set(true)
		scanProgress.Set(0.0)
		hostsDiscovered.Set([]HostRecord{})
		scannedIPCount.Set(0)
		openPortsTotal.Set(0)
		scanMutex.Unlock()

		go func() {
			defer func() {
				isScanning.Set(false)
				scanProgress.Set(1.0)
			}()

			startTime := time.Now()
			statusText.Set(fmt.Sprintf("Probing subnet %s across %s workers...", cidr, strconv.Itoa(concurrency)))

			ips := generateIPList(cidr)
			if len(ips) == 0 {
				statusText.Set("Error: Invalid subnet or CIDR notation.")
				return
			}
			totalIPCount.Set(len(ips))

			ports := parsePorts(portsStr)
			if len(ports) == 0 {
				ports = []int{80, 443, 22, 8080}
			}

			jobs := make(chan string, len(ips))
			for _, ip := range ips {
				jobs <- ip
			}
			close(jobs)

			var wg sync.WaitGroup
			var resultsMu sync.Mutex
			processed := 0
			total := len(ips)

			for w := 0; w < concurrency; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for ip := range jobs {
						select {
						case <-ctx.Done():
							return
						default:
						}

						res, alive := scanSingleHost(ip, ports, timeout)

						resultsMu.Lock()
						processed++
						scannedIPCount.Set(processed)
						scanProgress.Set(float64(processed) / float64(total))

						if alive {
							currentList := hostsDiscovered.Get()
							hostsDiscovered.Set(append(currentList, res))
							openPortsTotal.Update(func(c int) int {
								if res.OpenPorts != "Closed" && res.OpenPorts != "" {
									return c + len(strings.Split(res.OpenPorts, ", "))
								}
								return c
							})
						}
						resultsMu.Unlock()
					}
				}()
			}

			wg.Wait()
			elapsed := time.Since(startTime)
			aliveCount := len(hostsDiscovered.Get())
			statusText.Set(fmt.Sprintf("Scan completed in %v — %d live hosts discovered (%d open services).",
				elapsed.Round(time.Millisecond), aliveCount, openPortsTotal.Get()))
		}()
	}

	// 4. Native Menu Bar (Qt QMenuBar)
	topMenuBar := widgets.MenuBar(
		widgets.MenuBarItem{Title: "File"},
		widgets.MenuBarItem{Title: "Edit"},
		widgets.MenuBarItem{Title: "Scan"},
		widgets.MenuBarItem{Title: "View"},
		widgets.MenuBarItem{Title: "Tools"},
		widgets.MenuBarItem{Title: "Help"},
	)

	// 5. Main Action Toolbar (Qt QToolBar)
	mainToolbar := widgets.Toolbar(
		widgets.Button("Start Subnet Scan").OnClick(func() {
			timeout := time.Duration(timeoutMs.Get()) * time.Millisecond
			concurrency := int(concurrencyLimit.Get())
			startScan(targetSubnet.Get(), targetPorts.Get(), timeout, concurrency)
		}),

		widgets.Button("Scan Localhost").Secondary().OnClick(func() {
			targetSubnet.Set("127.0.0.1/32")
			startScan("127.0.0.1/32", "22,80,443,3000,8080,8000,9000,5432,3306,6379,27017", 100*time.Millisecond, 20)
		}),

		widgets.Button("Stop").Danger().OnClick(func() {
			scanMutex.Lock()
			if cancelScan != nil {
				cancelScan()
				statusText.Set("Scan interrupted by operator.")
			}
			isScanning.Set(false)
			scanMutex.Unlock()
		}),

		widgets.Button("Clear").Secondary().OnClick(func() {
			hostsDiscovered.Set([]HostRecord{})
			scannedIPCount.Set(0)
			openPortsTotal.Set(0)
			scanProgress.Set(0.0)
			statusText.Set("Session cleared.")
		}),

		widgets.Button("Export CSV").Outline(),
	)

	// 6. Left Profile / Subnet Navigator Sidebar
	sidebar := widgets.Sidebar("Network Explorer",
		widgets.SidebarItem{Title: "Active Subnet", Icon: "#", Selected: true},
		widgets.SidebarItem{Title: "Local Interfaces", Icon: "=", Badge: "eth0"},
		widgets.SidebarItem{Title: "DMZ Network", Icon: "$"},
		widgets.SidebarItem{Title: "Cloud VPC (AWS)", Icon: ">", Badge: "10.0.0.0"},
		widgets.SidebarItem{Title: "Saved Profiles (4)", Icon: "*"},
		widgets.SidebarItem{Title: "Port Presets", Icon: "≡"},
	)

	// 7. Qt-Style GroupBox: Scan Configuration
	configGroupBox := widgets.GroupBox("Subnet Target & Probe Parameters",
		ui.Column(
			ui.Row(
				widgets.TextField("192.168.1.0/24").
					WithLabel("Target CIDR / Subnet").
					Bind(targetSubnet),

				widgets.TextField("22, 80, 443, 8080, 3306, 5432").
					WithLabel("Port Probe List (CSV)").
					Bind(targetPorts),
			).GapSpacing(20),

			ui.Row(
				ui.Column(
					ui.Text("Probe Timeout: 200ms"),
					widgets.Slider(50, 1000).Bind(timeoutMs),
				).GapSpacing(4),

				ui.Column(
					ui.Text("Parallel Workers: 60 Threads"),
					widgets.Slider(10, 200).Bind(concurrencyLimit),
				).GapSpacing(4),
			).GapSpacing(24),
		).GapSpacing(10),
	)

	// 8. Live Telemetry Metric Cards
	metricsSection := ui.Row(
		widgets.Card("Active Devices",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("%d Online", len(hostsDiscovered.Get()))
				})).Size(20).Weight(700),
				widgets.Badge("Responding to Probe").Success(),
			).GapSpacing(4),
		),

		widgets.Card("Subnet Progress",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("%d / %d Hosts (%.0f%%)", scannedIPCount.Get(), totalIPCount.Get(), scanProgress.Get()*100)
				})).Size(20).Weight(700),
				widgets.Progress(scanProgress.Get()),
			).GapSpacing(4),
		),

		widgets.Card("Open Port Services",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("%d Services", openPortsTotal.Get())
				})).Size(20).Weight(700),
				widgets.Badge("TCP Verified").Info(),
			).GapSpacing(4),
		),
	).GapSpacing(12)

	// 9. Enterprise Qt Data Grid Table
	tableCols := []widgets.TableColumn{
		{Title: "IP Address", Width: 140, Field: "ip"},
		{Title: "Hostname / Reverse DNS", Width: 260, Field: "host"},
		{Title: "Latency", Width: 95, Field: "latency"},
		{Title: "Hardware Vendor", Width: 160, Field: "vendor"},
		{Title: "Status", Width: 95, Field: "status"},
		{Title: "Discovered Open Ports & Services", Width: 320, Field: "ports"},
	}

	tableData := func(row, col int) string {
		list := hostsDiscovered.Get()
		if row < 0 || row >= len(list) {
			return ""
		}
		item := list[row]
		switch col {
		case 0:
			return item.IP
		case 1:
			return item.Hostname
		case 2:
			return item.Latency
		case 3:
			return item.Vendor
		case 4:
			return item.Status
		case 5:
			return item.OpenPorts
		default:
			return ""
		}
	}

	gridGroupBox := widgets.GroupBox("Network Inventory & Live Host Discovery (QTableView)",
		ui.Column(
			widgets.Table(tableCols, len(hostsDiscovered.Get()), tableData),
		),
	)

	// 10. Bottom Status Bar (Qt QStatusBar)
	bottomStatusBar := widgets.StatusBar(statusText.Get(),
		widgets.StatusSegment{Text: "Subnet: " + targetSubnet.Get(), Width: 180},
		widgets.StatusSegment{Text: "Engine: Pure Go Async", Width: 160},
		widgets.StatusSegment{Text: "Latency: 1.1ms", Width: 120},
		widgets.StatusSegment{Text: "UTF-8", Width: 70},
	)

	// 11. Assemble Enterprise Workspace
	rightContent := ui.Padding(ui.All(14),
		ui.Column(
			configGroupBox,
			metricsSection,
			gridGroupBox,
		).GapSpacing(12),
	)

	centerSplit := widgets.SplitPane(widgets.SplitHorizontal, sidebar, rightContent)

	rootLayout := ui.Column(
		topMenuBar,
		mainToolbar,
		ui.Expanded(centerSplit),
		bottomStatusBar,
	)

	win.Content(rootLayout)

	fmt.Println("==========================================================")
	fmt.Println("🚀 Nova Enterprise Network Suite & IP Scanner Running...")
	fmt.Println("   • Design System:  Qt / Fusion Enterprise Slate")
	fmt.Println("   • Window Shell:   MenuBar + ToolBar + SplitPane + GridView + StatusBar")
	fmt.Println("==========================================================")

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
	}

	_ = win.SaveScreenshot("ip_scanner_preview.png")
	_ = selectedRowIndex
}

// Generate slice of IPs from CIDR string (e.g. 192.168.1.0/24 or single IP)
func generateIPList(cidr string) []string {
	var ips []string

	if !strings.Contains(cidr, "/") {
		if net.ParseIP(cidr) != nil {
			return []string{cidr}
		}
		cidr = cidr + "/24"
	}

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []string{"127.0.0.1"}
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		return ips[1 : len(ips)-1]
	}
	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func parsePorts(portsStr string) []int {
	var ports []int
	parts := strings.Split(portsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if val, err := strconv.Atoi(p); err == nil && val > 0 && val <= 65535 {
			ports = append(ports, val)
		}
	}
	return ports
}

// Scan a single IP host on given ports
func scanSingleHost(ip string, ports []int, timeout time.Duration) (HostRecord, bool) {
	var openPorts []string
	startPing := time.Now()
	var pingDuration time.Duration
	isAlive := false

	for _, port := range ports {
		target := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", target, timeout)
		if err == nil {
			_ = conn.Close()
			if !isAlive {
				pingDuration = time.Since(startPing)
				isAlive = true
			}
			openPorts = append(openPorts, strconv.Itoa(port))
		}
	}

	if !isAlive {
		return HostRecord{}, false
	}

	// Reverse DNS lookup
	hostname := "Unknown Device"
	names, err := net.LookupAddr(ip)
	if err == nil && len(names) > 0 {
		hostname = strings.TrimSuffix(names[0], ".")
	} else if ip == "127.0.0.1" {
		hostname = "localhost"
	}

	latencyStr := fmt.Sprintf("%.1f ms", float64(pingDuration.Microseconds())/1000.0)
	if pingDuration.Microseconds() == 0 {
		latencyStr = "< 1.0 ms"
	}

	portsStr := strings.Join(openPorts, ", ")
	if len(openPorts) == 0 {
		portsStr = "Responded (Filtered Ports)"
	}

	vendor := "Network Interface"
	if strings.HasSuffix(ip, ".1") {
		vendor = "Gateway / Cisco"
	} else if ip == "127.0.0.1" {
		vendor = "Loopback / Local"
	}

	return HostRecord{
		IP:        ip,
		Hostname:  hostname,
		Latency:   latencyStr,
		Vendor:    vendor,
		Status:    "Online",
		OpenPorts: portsStr,
	}, true
}

// Ensure forms import is retained
var _ = forms.Required
