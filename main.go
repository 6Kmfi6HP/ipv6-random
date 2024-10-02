package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	// 检查命令行参数
	if len(os.Args) != 4 {
		fmt.Println("Usage: ./main <ipv6/64> <count> <interface>")
		os.Exit(1)
	}

	ipv6Prefix := os.Args[1]
	count, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error: Invalid count parameter")
		os.Exit(1)
	}
	interfaceName := os.Args[3]

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
	if isPrivateIPv6(ip) {
		fmt.Println("Error: Private IPv6 addresses are not allowed")
		os.Exit(1)
	}

	// 生成并打印 IPv6/128 地址
	for i := 0; i < count; i++ {
		newIP := generateIPv6(ip)
		fmt.Printf("sudo ip addr add %s/128 dev %s;\n", newIP.String(), interfaceName)
	}
}

// generateIPv6 生成一个新的 IPv6 地址，保持前 64 位不变，随机生成后 64 位
func generateIPv6(baseIP net.IP) net.IP {
	newIP := make(net.IP, len(baseIP))
	copy(newIP, baseIP)

	// 随机生成后 64 位
	rand.Seed(time.Now().UnixNano())
	for i := 8; i < 16; i++ {
		newIP[i] = byte(rand.Intn(256))
	}

	return newIP
}

// isPrivateIPv6 检查给定的 IPv6 地址是否为私有地址
func isPrivateIPv6(ip net.IP) bool {
	// ULA (Unique Local Address) 范围: fc00::/7
	if ip[0] == 0xfc || ip[0] == 0xfd {
		return true
	}

	// Link-local 地址范围: fe80::/10
	if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
		return true
	}

	return false
}
