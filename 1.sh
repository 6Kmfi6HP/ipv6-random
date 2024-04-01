#!/bin/bash
# 创建目标目录，如果不存在的话
mkdir -p /usr/local/etc/sing-box

# 下载文件
wget -P /usr/local/etc/sing-box https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/cert.pem
wget -P /usr/local/etc/sing-box https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/private.key
wget -O /usr/local/etc/sing-box/config.json https://raw.githubusercontent.com/3Kmfi6HP/ipv6-random/main/singbox.json
