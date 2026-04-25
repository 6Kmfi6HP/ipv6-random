package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"
)

// XrayConfig 表示完整的 Xray 配置
type XrayConfig struct {
	Inbounds  []Inbound     `json:"inbounds"`
	Outbounds []Outbound    `json:"outbounds"`
	Routing   RoutingConfig `json:"routing"`
	Balancers BalancerRoot  `json:"balancers"`
	DNS       *DNSConfig    `json:"dns,omitempty"`
}

// Inbound 表示入站配置
type Inbound struct {
	Port     int             `json:"port"`
	Protocol string          `json:"protocol"`
	Tag      string          `json:"tag"`
	Settings InboundSettings `json:"settings"`
}

// InboundSettings 表示入站设置
type InboundSettings struct {
	Auth     string    `json:"auth,omitempty"`
	UDP      bool      `json:"udp,omitempty"`
	IP       string    `json:"ip,omitempty"`
	Accounts []Account `json:"accounts,omitempty"`
}

// Account 表示账户信息
type Account struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

// Outbound 表示出站配置
type Outbound struct {
	SendThrough string `json:"sendThrough,omitempty"`
	Protocol    string `json:"protocol"`
	Tag         string `json:"tag"`
}

// RoutingConfig 表示路由配置
type RoutingConfig struct {
	DomainStrategy string         `json:"domainStrategy,omitempty"`
	Rules          []RoutingRule  `json:"rules"`
	Balancers      []BalancerItem `json:"balancers"`
}

// RoutingRule 表示路由规则
type RoutingRule struct {
	Type        string   `json:"type"`
	InboundTag  string   `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
	IP          []string `json:"ip,omitempty"`
}

// BalancerItem 表示路由中的负载均衡器配置
type BalancerItem struct {
	Tag      string   `json:"tag"`
	Selector []string `json:"selector"`
}

// BalancerRoot 表示负载均衡器根配置
type BalancerRoot struct {
	Strategy BalancerStrategy `json:"strategy"`
}

// BalancerStrategy 表示负载均衡策略
type BalancerStrategy struct {
	Type string `json:"type"`
}

// DNSConfig 表示 DNS 配置
type DNSConfig struct {
	Hosts         map[string][]string `json:"hosts"`
	QueryStrategy string              `json:"queryStrategy"`
	Tag           string              `json:"tag"`
	Servers       []DNSServer         `json:"servers"`
}

// DNSServer 表示 DNS 服务器配置
type DNSServer struct {
	Address string `json:"address"`
	Port    int    `json:"port,omitempty"`
}

func main() {
	// 添加新的命令行标志
	verbose := flag.Bool("v", false, "Enable verbose logging")
	startPort := flag.Int("port", 20000, "Starting port number")
	socksUser := flag.String("user", "mute0857", "SOCKS username")
	socksPass := flag.String("pass", "Zxc13579", "SOCKS password")
	flag.Parse()

	// 设置日志输出
	if !*verbose {
		log.SetOutput(io.Discard)
	}

	// 检查命令行参数
	if flag.NArg() != 1 {
		fmt.Println("Usage: ./main [-v] <count>")
		os.Exit(1)
	}

	count, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Println("Error: Invalid count parameter")
		os.Exit(1)
	}

	// 自动获取主要上网接口和IPv6地址
	interfaceName, ipv6Prefix, err := getInterfaceAndIPv6()
	log.Println("Interface Name:", interfaceName)
	log.Println("IPv6 Prefix:", ipv6Prefix)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// 解析 IPv6 地址和前缀长度
	ip, ipNet, err := net.ParseCIDR(ipv6Prefix)
	if err != nil {
		fmt.Println("Error: Invalid IPv6 address")
		os.Exit(1)
	}

	// 确保输入的是 /64 前缀
	if ones, _ := ipNet.Mask.Size(); ones != 64 {
		fmt.Println("Error: Input must be a /64 IPv6 prefix")
		os.Exit(1)
	}

	// 检查是否为局域网 IPv6 地址
	if isPrivateIPv6(ipv6Prefix) {
		fmt.Println("Error: Private IPv6 addresses are not allowed")
		os.Exit(1)
	}

	// 生成并打印 IPv6/128 地址
	var allAddresses []string
	for i := 0; i < count; i++ {
		newIP := generateIPv6(ip)
		ipCmd := fmt.Sprintf("sudo ip addr add %s/128 dev %s", newIP.String(), interfaceName)
		fmt.Println(ipCmd + ";")
		allAddresses = append(allAddresses, newIP.String())
	}

	// 生成配置文件
	if err := generateXrayConfig(allAddresses, interfaceName, *startPort, *socksUser, *socksPass); err != nil {
		fmt.Printf("Error generating config: %v\n", err)
		os.Exit(1)
	}
}

// generateIPv6 生一个新的 IPv6 地址，保持前 64 位不变，随机生成后 64 位
func generateIPv6(baseIP net.IP) net.IP {
	newIP := make(net.IP, len(baseIP))
	copy(newIP, baseIP)

	// 随机生成后 64 位
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 8; i < 16; i++ {
		newIP[i] = byte(r.Intn(256))
	}

	return newIP
}

// isPrivateIPv6 检查给定的 IPv6 地址是否为私有地址
func isPrivateIPv6(ipStr string) bool {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return false
	}

	// ULA (Unique Local Address) 范围: fc00::/7
	if ip.Is6() && (ip.AsSlice()[0] == 0xfc || ip.AsSlice()[0] == 0xfd) {
		return true
	}

	// Link-local 地址范围: fe80::/10
	if ip.Is6() && ip.AsSlice()[0] == 0xfe && (ip.AsSlice()[1]&0xc0) == 0x80 {
		return true
	}

	return false
}

// 修改后的函数：自动获取主要上网接口和IPv6地址
func getInterfaceAndIPv6() (string, string, error) {
	iface, err := getDefaultInterface()
	if err != nil {
		return "", "", fmt.Errorf("failed to get default interface: %v", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", "", fmt.Errorf("failed to get addresses for interface %s: %v", iface.Name, err)
	}

	var globalIPv6 net.IP
	var linkLocalIPv6 net.IP

	for _, addr := range addrs {
		ipStr := strings.Split(addr.String(), "/")[0] // Remove the subnet mask
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.To4() == nil { // This is an IPv6 address
			if !ip.IsLoopback() {
				if !ip.IsLinkLocalUnicast() {
					globalIPv6 = ip
					break // We found a global IPv6, no need to continue
				} else if linkLocalIPv6 == nil {
					linkLocalIPv6 = ip
				}
			}
		}
	}

	if globalIPv6 != nil {
		return iface.Name, globalIPv6.String() + "/64", nil
	} else if linkLocalIPv6 != nil {
		return iface.Name, linkLocalIPv6.String() + "/64", nil
	}

	return "", "", fmt.Errorf("no suitable IPv6 address found on interface %s", iface.Name)
}

// 更新的函数：获取默认路由接口
func getDefaultInterface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	log.Println("Checking interfaces:")
	for _, iface := range interfaces {
		log.Printf("Interface: %s, Flags: %v", iface.Name, iface.Flags)

		if iface.Flags&net.FlagUp == 0 {
			log.Println("  Skipping: Interface is down")
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			log.Println("  Skipping: Loopback interface")
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("  Error getting addresses: %v", err)
			continue
		}

		log.Printf("  Addresses:")
		for _, addr := range addrs {
			log.Printf("    %s", addr.String())
			ipStr := strings.Split(addr.String(), "/")[0] // Remove the subnet mask
			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Println("      Failed to parse IP address")
				continue
			}
			if ip.To4() != nil && !ip.IsLoopback() && !ip.IsPrivate() {
				log.Println("      Found suitable public IPv4 address")
				return &iface, nil
			}
			if ip.To4() == nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
				log.Println("      Found suitable public IPv6 address")
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no default interface found")
}

func generateXrayConfig(ipAddresses []string, _ string, startPort int, socksUser, socksPass string) error {
	config := XrayConfig{
		Inbounds: []Inbound{
			{
				Port:     startPort,
				Protocol: "socks",
				Tag:      "tag_all",
				Settings: InboundSettings{
					Auth: "password",
					UDP:  true,
					Accounts: []Account{{
						User: socksUser,
						Pass: socksPass,
					}},
				},
			},
		},
		Outbounds: make([]Outbound, 0, len(ipAddresses)+1),
		Routing: RoutingConfig{
			Rules: []RoutingRule{{
				Type:        "field",
				InboundTag:  "tag_all",
				BalancerTag: "balancer",
			}},
			Balancers: []BalancerItem{{
				Tag:      "balancer",
				Selector: make([]string, 0, len(ipAddresses)),
			}},
		},
		Balancers: BalancerRoot{
			Strategy: BalancerStrategy{
				Type: "roundRobin",
			},
		},
	}

	// 为每个IP添加出站配置
	selectors := make([]string, 0, len(ipAddresses))
	for i, ip := range ipAddresses {
		tag := fmt.Sprintf("tag_%d", i+1)
		selectors = append(selectors, tag)

		config.Outbounds = append(config.Outbounds, Outbound{
			SendThrough: ip,
			Protocol:    "freedom",
			Tag:         tag,
		})
	}

	// 设置负载均衡选择器
	config.Routing.Balancers[0].Selector = selectors

	// 添加默认出站
	config.Outbounds = append(config.Outbounds, Outbound{
		Protocol: "freedom",
		Tag:      "tag_all",
	})

	// 写入配置文件
	f, err := os.Create("config.json")
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	return nil
}
