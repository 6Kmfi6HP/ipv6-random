#!/bin/bash
DEFAULT_START_PORT=20000                         #默认起始端口
DEFAULT_SOCKS_USERNAME="mute0857"                #默认socks账号
DEFAULT_SOCKS_PASSWORD="Zxc13579"                #默认socks密码
DEFAULT_WS_PATH="/ws"                            #默认ws路径
DEFAULT_UUID=$(cat /proc/sys/kernel/random/uuid) #默认随机UUID

# 显示 ip a 内容
echo "当前网络接口信息："
ip a

# 下载 ipv6gen
download_ipv6gen() {
    local arch=$(uname -m)
    local url=""

    case $arch in
    x86_64)
        url="https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/build/ipv6gen_linux_amd64"
        ;;
    aarch64)
        url="https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/build/ipv6gen_linux_arm64"
        ;;
    i386 | i686)
        url="https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/build/ipv6gen_linux_386"
        ;;
    *)
        echo "不支持的架构: $arch"
        exit 1
        ;;
    esac

    echo "下载 ipv6gen..."
    wget -O ipv6gen $url
    chmod +x ipv6gen
}

# 检查 ipv6gen 是否存在，如果不存在则下载
if [ ! -f "ipv6gen" ]; then
    download_ipv6gen
else
    echo "ipv6gen 已存在，跳过下载"
fi

# 用户输入 IPv6 前缀、地址数量和接口名称
read -p "请输入 IPv6 前缀（格式：xxxx:xxxx:xxxx:xxxx::/64）: " IPV6_PREFIX
read -p "请输入要生成的 IPv6 地址数量: " NUM_ADDRESSES
read -p "请输入网络接口名称: " INTERFACE_NAME

# 生成添加 IP 的命令
echo "生成添加 IP 的命令..."
ADD_IP_COMMANDS=$(./ipv6gen "$IPV6_PREFIX" "$NUM_ADDRESSES" "$INTERFACE_NAME")
echo "$ADD_IP_COMMANDS"

# 停止和禁用systemd-resolved服务
echo "停止和禁用systemd-resolved服务..."
sudo systemctl stop systemd-resolved.service
echo "禁用systemd-resolved服务..."
sudo systemctl disable systemd-resolved.service

# 设置DNS服务器
echo "设置DNS服务器..."
# 备份原始resolv.conf文件
sudo cp /etc/resolv.conf /etc/resolv.conf.backup
# 创建新的resolv.conf文件
sudo tee /etc/resolv.conf >/dev/null <<EOT
nameserver 1.1.1.1
nameserver 8.8.8.8
EOT


# 执行添加 IP 的命令
echo "执行添加 IP 的命令..."
eval "$ADD_IP_COMMANDS"
echo "添加 IP 完成."
# 获取所有 IP 地址
IP_ADDRESSES=($(hostname -I))

# 安装 xray
install_xray() {
    echo "安装 Xray..."
    # Install unzip using the appropriate package manager
    if command -v apt-get &>/dev/null; then
        apt-get install unzip -y
    elif command -v yum &>/dev/null; then
        yum install unzip -y
    else
        echo "无法安装 unzip。请手动安装后重试。"
        exit 1
    fi

    # Download and install Xray
    wget https://github.com/XTLS/Xray-core/releases/download/v1.8.24/Xray-linux-64.zip
    unzip Xray-linux-64.zip
    mv xray /usr/local/bin/xrayK
    chmod +x /usr/local/bin/xrayK

    # Create xrayK service file
    cat <<EOF >/etc/systemd/system/xrayK.service
[Unit]
Description=xrayK Service
After=network.target

[Service]
ExecStart=/usr/local/bin/xrayK -c /etc/xrayK/config.toml
Restart=on-failure
User=nobody
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd, enable and start xrayK service
    systemctl daemon-reload
    systemctl enable xrayK.service
    systemctl start xrayK.service
    echo "Xray 安装完成."

    # Clean up downloaded files
    rm -f Xray-linux-64.zip
}

# 配置 xray
config_xray() {
    config_type=$1
    mkdir -p /etc/xrayK
    if [ "$config_type" != "socks" ] && [ "$config_type" != "vmess" ]; then
        echo "类型错误！仅支持socks和vmess."
        exit 1
    fi

    read -p "起始端口 (默认 $DEFAULT_START_PORT): " START_PORT
    START_PORT=${START_PORT:-$DEFAULT_START_PORT}
    if [ "$config_type" == "socks" ]; then
        read -p "SOCKS 账号 (默认 $DEFAULT_SOCKS_USERNAME): " SOCKS_USERNAME
        SOCKS_USERNAME=${SOCKS_USERNAME:-$DEFAULT_SOCKS_USERNAME}

        read -p "SOCKS 密码 (默认 $DEFAULT_SOCKS_PASSWORD): " SOCKS_PASSWORD
        SOCKS_PASSWORD=${SOCKS_PASSWORD:-$DEFAULT_SOCKS_PASSWORD}
    elif [ "$config_type" == "vmess" ]; then
        read -p "UUID (默认随机): " UUID
        UUID=${UUID:-$DEFAULT_UUID}
        read -p "WebSocket 路径 (默认 $DEFAULT_WS_PATH): " WS_PATH
        WS_PATH=${WS_PATH:-$DEFAULT_WS_PATH}
    fi

    for ((i = 0; i < ${#IP_ADDRESSES[@]}; i++)); do
        config_content+="[[inbounds]]\n"
        config_content+="port = $((START_PORT + i))\n"
        config_content+="protocol = \"$config_type\"\n"
        config_content+="tag = \"tag_$((i + 1))\"\n"
        config_content+="[inbounds.settings]\n"
        if [ "$config_type" == "socks" ]; then
            config_content+="auth = \"password\"\n"
            config_content+="udp = true\n"
            config_content+="ip = \"${IP_ADDRESSES[i]}\"\n"
            config_content+="[[inbounds.settings.accounts]]\n"
            config_content+="user = \"$SOCKS_USERNAME\"\n"
            config_content+="pass = \"$SOCKS_PASSWORD\"\n"
        elif [ "$config_type" == "vmess" ]; then
            config_content+="[[inbounds.settings.clients]]\n"
            config_content+="id = \"$UUID\"\n"
            config_content+="[inbounds.streamSettings]\n"
            config_content+="network = \"ws\"\n"
            config_content+="[inbounds.streamSettings.wsSettings]\n"
            config_content+="path = \"$WS_PATH\"\n\n"
        fi
        config_content+="[[outbounds]]\n"
        config_content+="sendThrough = \"${IP_ADDRESSES[i]}\"\n"
        config_content+="protocol = \"freedom\"\n"
        config_content+="tag = \"tag_$((i + 1))\"\n\n"
        config_content+="[[routing.rules]]\n"
        config_content+="type = \"field\"\n"
        config_content+="inboundTag = \"tag_$((i + 1))\"\n"
        config_content+="outboundTag = \"tag_$((i + 1))\"\n\n\n"
    done
    # balancers config
    config_content+="[[inbounds]]\n"
    config_content+="port = $((START_PORT - 50))\n"
    config_content+="protocol = \"socks\"\n"
    config_content+="tag = \"tag_all\"\n"
    config_content+="\n"
    config_content+="[inbounds.settings]\n"
    config_content+="auth = \"password\"\n"
    config_content+="udp = true\n"
    config_content+="\n"
    config_content+="[[inbounds.settings.accounts]]\n"
    config_content+="user = \"$SOCKS_USERNAME\"\n"
    config_content+="pass = \"$SOCKS_PASSWORD\"\n"
    config_content+="\n"
    config_content+="[[routing.rules]]\n"
    config_content+="type = \"field\"\n"
    config_content+="ip = [\"::/0\"]\n"
    config_content+="balancerTag = \"balancer\"\n"
    config_content+="inboundTag = \"tag_all\"\n"
    config_content+="\n"
    config_content+="[[routing.balancers]]\n"
    config_content+="tag = \"balancer\"\n"
    config_content+="selector = [\n"
    for ((i = 0; i < ${#IP_ADDRESSES[@]}; i++)); do
        config_content+="    \"tag_$((i + 1))\",\n"
    done
    config_content+="]\n"
    config_content+="\n"
    config_content+="[balancers.strategy]\n"
    config_content+="type = \"roundRobin\"\n"
    config_content+="\n"
    config_content+="[[outbounds]]\n"
    config_content+="protocol = \"freedom\"\n"
    config_content+="tag = \"tag_all\"\n"

    # Add DNS configuration
    config_content+="[dns]\n"
    config_content+="hosts = { \"dns.google\" = [\"8.8.8.8\", \"8.8.4.4\"] }\n"
    config_content+="queryStrategy = \"UseIP\"\n"
    config_content+="tag = \"dns_inbound\"\n\n"
    config_content+="[[dns.servers]]\n"
    config_content+="address = \"8.8.8.8\"\n"
    config_content+="port = 53\n\n"
    config_content+="[[dns.servers]]\n"
    config_content+="address = \"1.1.1.1\"\n"
    config_content+="port = 53\n\n"
    config_content+="[[dns.servers]]\n"
    config_content+="address = \"https://dns.google/dns-query\"\n\n"

    # Add routing configuration
    config_content+="[routing]\n"
    config_content+="domainStrategy = \"IPOnDemand\"\n\n"

    # Add routing rule for IPv4
    config_content+="[[routing.rules]]\n"
    config_content+="type = \"field\"\n"
    config_content+="ip = [\"0.0.0.0/0\"]\n"
    config_content+="inboundTag = [\"dns_inbound\"]\n"
    config_content+="outboundTag = \"direct\"\n\n"

    echo -e "$config_content" >/etc/xrayK/config.toml
    systemctl restart xrayK.service
    systemctl --no-pager status xrayK.service
    echo ""
    echo "生成 $config_type 配置完成"
    echo "起始端口:$START_PORT"
    echo "结束端口:$(($START_PORT + $i - 1))"
    echo "随机端口:$(($START_PORT - 50))"
    if [ "$config_type" == "socks" ]; then
        echo "socks账号:$SOCKS_USERNAME"
        echo "socks密码:$SOCKS_PASSWORD"
    elif [ "$config_type" == "vmess" ]; then
        echo "UUID:$UUID"
        echo "ws路径:$WS_PATH"
    fi
    echo ""
}
check_ipv6_valid() {
    # 使用代理检查IP信息
    echo "使用ipinfo.io检查IP信息(ipv4)："
    curl --proxy socks5h://mute0857:Zxc13579@127.0.0.1:$((START_PORT - 50)) ipinfo.io

    echo -e "\n使用ip.sb检查IP(ipv6)："
    curl --proxy socks5h://mute0857:Zxc13579@127.0.0.1:$((START_PORT - 50)) ip.sb

    echo -e "\n脚本执行完毕。"
}

main() {
    [ -x "$(command -v xrayK)" ] || install_xray
    if [ $# -eq 1 ]; then
        config_type="$1"
    else
        read -p "选择生成的节点类型 (socks/vmess): " config_type
    fi
    if [ "$config_type" == "vmess" ]; then
        config_xray "vmess"
    elif [ "$config_type" == "socks" ]; then
        config_xray "socks"
    else
        echo "未正确选择类型 使用默认sokcs配置."
        config_xray "socks"
    fi
    check_ipv6_valid
}
main "$@"
