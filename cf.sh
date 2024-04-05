#!/bin/bash

target_ip=$1
PORTS=(80 8080 8880 2052 2082 2086 2095 443 2053 2083 2087 2096 8443)

echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
sysctl -p

for PORT in "${PORTS[@]}"; do
    iptables -t nat -A PREROUTING -p tcp --dport $PORT -j DNAT --to $target_ip:$PORT
    iptables -t nat -A POSTROUTING -p tcp -d $target_ip --dport $PORT -j MASQUERADE
    iptables -A FORWARD -p tcp -d $target_ip --dport $PORT -j ACCEPT
done

iptables -P FORWARD DROP

echo "iptables 规则以及IP转发已更新。"
