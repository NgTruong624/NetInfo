<!-- e5560857-3b55-4619-a14c-95282c3f492f 13fd7f66-abfa-4fd0-be25-60d13f0ed6e6 -->
# Kế hoạch thực thi NetInfo (Go, module ở root)

## Cấu trúc dự án (root repo)

```
NetInfo/
├── go.mod
├── go.sum
├── main.go                  # Entry point, menu chính
├── cmd/
│   └── root.go              # Command/menu wiring
├── network/
│   ├── interfaces.go        # Thông tin interfaces
│   ├── ip.go                # Local/Public IP
│   ├── dns.go               # DNS servers
│   ├── gateway.go           # Default gateway
│   ├── routes.go            # Routing table
│   ├── connections.go       # Active connections
│   └── ping.go              # Ping utility
├── display/
│   ├── menu.go              # Interactive menu
│   ├── formatter.go         # Bảng/formatting
│   └── colors.go            # Màu terminal
└── utils/
    ├── platform.go          # Phát hiện nền tảng
    └── helpers.go           # Helpers chung
```

## Công nghệ & Dependencies

- Go 1.21+
- github.com/shirou/gopsutil/v3
- github.com/manifoldco/promptui
- github.com/fatih/color
- github.com/olekukonko/tablewriter
- Thư viện chuẩn: net, os/exec, runtime, fmt, time, context

## Quy ước nền tảng (tối ưu & fallback)

- Windows (ưu tiên PowerShell):
  - DNS: `Get-DnsClientServerAddress -AddressFamily IPv4, IPv6`
  - Gateway/Routes: `Get-NetRoute` (fallback `route print`)
  - Ping: `ping -n <count> <host>`
- Linux:
  - DNS: đọc `/etc/resolv.conf`
  - Gateway/Routes: `ip -j route` (fallback parse `ip route` hoặc `/proc/net/route`)
  - Ping: `ping -c <count> <host>`
- Connections: `gopsutil/net.Connections()`; process name khi có quyền. Graceful degrade nếu thiếu quyền.
- Public IP: gọi tuần tự kèm timeout 2s mỗi endpoint: `https://api.ipify.org`, `https://ifconfig.me`, `https://icanhazip.com`.

## Dòng chảy chính

1) Kiểm tra nền tảng (`utils/platform.go`)

2) Hiển thị banner, menu (`display/menu.go` + `promptui`)

3) Thực thi chức năng theo lựa chọn và format bảng (`display/formatter.go` + `tablewriter`, `color`)

4) Pausing và quay lại menu; xử lý lỗi thân thiện, gợi ý quyền nếu cần

## Phác thảo triển khai tệp chính

- `utils/platform.go`: hàm `IsWindows()`, `IsLinux()`, `CommandWithTimeout(ctx, name, args...)`
- `network/dns.go` (Windows sample lệnh):
```9:18:network/dns.go
// Windows: powershell -NoProfile -Command Get-DnsClientServerAddress -AddressFamily IPv4,IPv6 | ConvertTo-Json
```

- `network/gateway.go` (Linux ưu tiên JSON):
```9:14:network/gateway.go
// Linux: ip -j route show default | jq-less parse JSON (go: exec ip -j route, unmarshal JSON)
```

- `network/ip.go`: local IP từ `net.Interfaces()`; public IP qua HTTP client với timeout
- `display/formatter.go`: helpers render bảng, truncation cột dài
- `main.go`: vòng lặp menu, gọi handlers theo case

## Build & Phát hành

- Build hiện tại: `go build -o netinfo` (hoặc `.exe` trên Windows)
- Cross-compile: `GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o netinfo.exe`, tương tự cho Linux
- (Tùy chọn) Thiết lập `goreleaser` sau

## Kiểm thử

- Manual trên Windows 10/11 và Ubuntu 20.04+
- Kiểm tra các tình huống: không có mạng, VPN, thiếu quyền admin/root
- Unit nhẹ cho helpers/formatters (nếu hợp lý trong thời gian)

## Tài liệu

- README: giới thiệu, cài đặt, quyền cần thiết theo tính năng, hướng dẫn sử dụng (ảnh màn hình), build, troubleshooting

### To-dos

- [x] Khởi tạo module Go ở root và scaffold thư mục
- [x] Thêm và vendor dependencies (gopsutil, promptui, color, tablewriter)
- [x] Implement utils/platform.go và utils/helpers.go
- [x] Tạo display/menu, colors, formatter với tablewriter
- [x] Hiển thị network interfaces (IP, MAC, MTU, status)
- [x] Lấy local IPs và public IP với fallback endpoints
- [x] Lấy DNS servers (PowerShell/resolve.conf)
- [x] Lấy default gateway (Get-NetRoute/ip -j route)
- [x] Hiển thị routing table (PS/ip -j route)
- [x] Hiển thị active connections và process khi có quyền
- [ ] Ping utility theo nền tảng với parse kết quả
- [ ] Tích hợp tất cả chức năng vào menu và main.go
- [ ] Chuẩn hóa error handling, timeouts, degrade thông báo
- [ ] Test thủ công trên Windows và Ubuntu, fix lỗi
- [ ] Viết README đầy đủ hướng dẫn và ảnh chụp