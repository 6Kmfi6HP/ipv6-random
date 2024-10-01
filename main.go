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

	// 生成并打印 IPv6/128 地址
	for i := 0; i < count; i++ {
		newIP := generateIPv6(ip, i)
		fmt.Printf("sudo ip addr add %s/128 dev %s;\n", newIP.String(), interfaceName)
	}
}

func generateIPv6(baseIP net.IP, _ int) net.IP {
	newIP := make(net.IP, len(baseIP))
	copy(newIP, baseIP)

	// 随机生成后 64 位
	rand.Seed(time.Now().UnixNano())
	for i := 8; i < 16; i++ {
		newIP[i] = byte(rand.Intn(256))
	}

	return newIP
}
