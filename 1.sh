#!/bin/bash
# 创建目标目录，如果不存在的话
mkdir -p /usr/local/etc/sing-box

# 下载文件
wget -P /usr/local/etc/sing-box https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/cert.pem
wget -P /usr/local/etc/sing-box https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/private.key
wget -O /usr/local/etc/sing-box/config.json https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/singbox.json
wget -O /etc/s-box/sb.json https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/singbox.json

sudo apt-get install netfilter-persistent

# IPv4
iptables -t nat -A PREROUTING -i eth0 -p udp --dport 50001:59999 -j DNAT --to-destination :55555
# IPv6
ip6tables -t nat -A PREROUTING -i eth0 -p udp --dport 50001:59999 -j DNAT --to-destination :55555
netfilter-persistent save >/dev/null 2>&1

# IPv4
iptables -t nat -A PREROUTING -i eth0 -p udp --dport 50001:59999 -j DNAT --to-destination 216.73.156.201:55555
# IPv6
ip6tables -t nat -A PREROUTING -i eth0 -p udp --dport 50001:59999 -j DNAT --to-destination 216.73.156.201:55555
netfilter-persistent save >/dev/null 2>&1
