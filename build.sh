#!/bin/bash

# 设置 Go 程序的名称
APP_NAME="ipv6gen"

# 设置源文件
SOURCE_FILE="main.go"

# 设置下载链接前缀
DOWNLOAD_PREFIX="https://raw.githubusercontent.com/6Kmfi6HP/ipv6-random/main/"

# 定义要编译的目标平台
PLATFORMS=("windows/amd64" "windows/386" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "linux/386")

# 创建输出目录
mkdir -p build

# 创建下载链接文件
LINK_FILE="download_links.txt"
echo "Download Links:" > $LINK_FILE

# 遍历所有平台并编译
for PLATFORM in "${PLATFORMS[@]}"
do
    # 分割平台字符串为操作系统和架构
    IFS='/' read -r -a array <<< "$PLATFORM"
    GOOS=${array[0]}
    GOARCH=${array[1]}
    
    # 设置输出文件名
    if [ $GOOS = "windows" ]; then
        OUTPUT_NAME=$APP_NAME'_'$GOOS'_'$GOARCH'.exe'
    else
        OUTPUT_NAME=$APP_NAME'_'$GOOS'_'$GOARCH
    fi
    
    # 编译
    echo "Compiling for $GOOS $GOARCH"
    env GOOS=$GOOS GOARCH=$GOARCH go build -o build/$OUTPUT_NAME $SOURCE_FILE
    
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    
    # 生成下载链接并添加到文件
    DOWNLOAD_LINK="${DOWNLOAD_PREFIX}build/${OUTPUT_NAME}"
    echo "$GOOS $GOARCH: $DOWNLOAD_LINK" >> $LINK_FILE
done

echo "Build completed successfully."
echo "Download links have been saved to $LINK_FILE"