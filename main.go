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
)

func main() {
	// 添加新的命令行标志
	verbose := flag.Bool("v", false, "Enable verbose logging")
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
	for i := 0; i < count; i++ {
		newIP := generateIPv6(ip)
		fmt.Printf("sudo ip addr add %s/128 dev %s;\n", newIP.String(), interfaceName)
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
