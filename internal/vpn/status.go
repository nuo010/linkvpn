package vpn

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"singleOpenVpn/internal/model"
)

// ClientUsage 单条连接使用统计（与使用统计界面指标对应）
type ClientUsage struct {
	CommonName   string `json:"common_name"`   // 用户/客户端名
	RealIP       string `json:"real_ip"`       // 用户登录 IP（Real Address 去掉端口）
	VirtualIP    string `json:"virtual_ip"`    // 用户 DHCP/VPN IP
	BytesRecv    int64  `json:"bytes_recv"`    // 服务器收到字节（用户上传流量）
	BytesSent    int64  `json:"bytes_sent"`    // 服务器发出字节（用户下载流量）
	ConnectedAt  string `json:"connected_at"`  // 上线时间
	DurationSecs int64  `json:"duration_secs"` // 在线时长（秒）
}

// ParseStatusFile 解析 OpenVPN status 文件，返回当前连接列表
// 格式参考：OpenVPN CLIENT LIST / Common Name,Real Address,Virtual Address,Bytes Received,Bytes Sent,Connected Since
func ParseStatusFile(path string) ([]ClientUsage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var list []ClientUsage
	sc := bufio.NewScanner(f)
	var header []string
	var hasVirtualInHeader bool

	inRouting := false
	var routingHeader []string
	virtualByCN := make(map[string]string)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "OpenVPN CLIENT LIST") {
			header = nil
			inRouting = false
			continue
		}
		if strings.HasPrefix(line, "ROUTING TABLE") {
			inRouting = true
			header = nil
			continue
		}
		if inRouting {
			if strings.HasPrefix(line, "Virtual Address,") {
				routingHeader = strings.Split(line, ",")
				for i := range routingHeader {
					routingHeader[i] = strings.TrimSpace(routingHeader[i])
				}
				continue
			}
			parts := splitStatusLine(line)
			if len(routingHeader) > 0 && len(parts) >= 2 {
				var vAddr, cn string
				for i, h := range routingHeader {
					if i >= len(parts) {
						break
					}
					p := strings.TrimSpace(parts[i])
					if h == "Virtual Address" {
						vAddr = p
					}
					if h == "Common Name" {
						cn = p
					}
				}
				if cn != "" && vAddr != "" {
					virtualByCN[cn] = vAddr
				}
			}
			continue
		}
		if strings.HasPrefix(line, "Common Name,") {
			header = strings.Split(line, ",")
			for i, h := range header {
				header[i] = strings.TrimSpace(h)
			}
			hasVirtualInHeader = false
			for _, h := range header {
				if h == "Virtual Address" {
					hasVirtualInHeader = true
					break
				}
			}
			continue
		}
		if len(header) == 0 {
			continue
		}
		parts := splitStatusLine(line)
		if len(parts) < 5 {
			continue
		}
		commonName := ""
		realAddr := ""
		virtualIP := ""
		var bytesRecv, bytesSent int64
		connectedSince := ""

		for i, p := range parts {
			if i >= len(header) {
				break
			}
			switch header[i] {
			case "Common Name":
				commonName = p
			case "Real Address":
				realAddr = p
			case "Virtual Address":
				virtualIP = p
			case "Bytes Received":
				bytesRecv, _ = strconv.ParseInt(p, 10, 64)
			case "Bytes Sent":
				bytesSent, _ = strconv.ParseInt(p, 10, 64)
			case "Connected Since":
				connectedSince = p
			}
		}
		if !hasVirtualInHeader && len(parts) >= 5 {
			commonName = parts[0]
			realAddr = parts[1]
			bytesRecv, _ = strconv.ParseInt(parts[2], 10, 64)
			bytesSent, _ = strconv.ParseInt(parts[3], 10, 64)
			connectedSince = parts[4]
		}

		realIP, _ := splitHostPort(realAddr)
		if realIP == "" {
			realIP = realAddr
		}

		connectedSince = strings.TrimSpace(connectedSince)
		connectedAt := parseConnectedSince(connectedSince)
		dur := int64(0)
		if !connectedAt.IsZero() {
			dur = int64(time.Since(connectedAt).Seconds())
		}

		list = append(list, ClientUsage{
			CommonName:   commonName,
			RealIP:       realIP,
			VirtualIP:    virtualIP,
			BytesRecv:    bytesRecv,
			BytesSent:    bytesSent,
			ConnectedAt:  connectedSince,
			DurationSecs: dur,
		})
	}
	// 用 ROUTING TABLE 中解析到的 Virtual Address 补全
	for i := range list {
		if list[i].VirtualIP == "" && list[i].CommonName != "" {
			if v, ok := virtualByCN[list[i].CommonName]; ok {
				list[i].VirtualIP = v
			}
		}
	}
	return list, sc.Err()
}

// ParseConnectedSince 解析 OpenVPN status 中 Connected Since 多种格式（供 handler 等复用）
func ParseConnectedSince(s string) time.Time {
	return parseConnectedSince(s)
}

func parseConnectedSince(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	loc := model.Shanghai
	// 无时区后缀的一律按东八区墙钟解析
	localLayouts := []string{
		"Mon Jan _2 15:04:05 2006",
		"Mon Jan 2 15:04:05 2006",
		"2 Jan 2006 15:04:05",
		"_2 Jan 2006 15:04:05",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
	}
	for _, layout := range localLayouts {
		if t, err := time.ParseInLocation(layout, s, loc); err == nil {
			return t
		}
	}
	// 显式 Z 的按 UTC 解析再保留瞬时（用于 duration）；若你希望也当东八区墙钟可改为 ParseInLocation 去掉 Z
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", s); err == nil {
		return t
	}
	return time.Time{}
}

func splitStatusLine(line string) []string {
	var parts []string
	var cur strings.Builder
	for _, r := range line {
		if r == ',' {
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteRune(r)
	}
	parts = append(parts, strings.TrimSpace(cur.String()))
	return parts
}

func splitHostPort(addr string) (host, port string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", ""
	}
	return host, port
}

// FindOpenVPNPID 尝试在 /proc 中查找正在运行的 openvpn 进程，返回第一个匹配的 PID。
// 仅在 Linux 环境（包含 /proc）下有效；失败时返回 (0, false)。
func FindOpenVPNPID() (int, bool) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 0 {
			continue
		}
		data, err := os.ReadFile("/proc/" + e.Name() + "/cmdline")
		if err != nil || len(data) == 0 {
			continue
		}
		cmdline := string(data)
		if strings.Contains(cmdline, "openvpn") {
			return pid, true
		}
	}
	return 0, false
}

var (
	openvpnVersionOnce sync.Once
	openvpnVersion     string
)

// GetOpenVPNVersion 返回 openvpn --version 的第一行（带版本号），失败返回空串。
// 为了避免频繁 exec，这里做了进程内缓存（只取一次）。
func GetOpenVPNVersion() string {
	openvpnVersionOnce.Do(func() {
		bin, err := exec.LookPath("openvpn")
		if err != nil || bin == "" {
			// 非交互式启动时 PATH 可能不含 /usr/sbin 等目录，兜底常见位置
			for _, p := range []string{
				"/usr/sbin/openvpn",
				"/usr/local/sbin/openvpn",
				"/usr/bin/openvpn",
				"/usr/local/bin/openvpn",
			} {
				if _, statErr := os.Stat(p); statErr == nil {
					bin = p
					break
				}
			}
		}
		if bin == "" {
			return
		}

		out, err := exec.Command(bin, "--version").Output()
		if err != nil || len(out) == 0 {
			return
		}
		// 例：OpenVPN 2.6.8 x86_64-pc-linux-gnu [SSL (OpenSSL)] ...
		line := bytes.SplitN(out, []byte{'\n'}, 2)[0]
		openvpnVersion = parseOpenVPNVersionLine(strings.TrimSpace(string(line)))
	})
	return openvpnVersion
}

func parseOpenVPNVersionLine(line string) string {
	// 典型输出：
	// OpenVPN 2.6.3 x86_64-pc-linux-gnu [SSL (OpenSSL)] ...
	// 我们只展示版本号（2.6.3），避免 UI 过长。
	fields := strings.Fields(line)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "openvpn") {
		return fields[1]
	}
	// 兜底：无法解析时返回整行（至少有信息可用）
	return line
}
