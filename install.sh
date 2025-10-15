#!/bin/bash

# 企业微信通知服务安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/GhostLee/wecom-notifier/main/install.sh | bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 检测操作系统和架构
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            error "不支持的操作系统: $OS"
            ;;
    esac

    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            error "不支持的架构: $ARCH"
            ;;
    esac

    info "检测到平台: ${OS}-${ARCH}"
}

# 获取最新版本
get_latest_version() {
    info "获取最新版本信息..."
    LATEST_VERSION=$(curl -fsSL https://api.github.com/repos/GhostLee/wecom-notifier/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_VERSION" ]; then
        error "无法获取最新版本信息"
    fi
    
    info "最新版本: $LATEST_VERSION"
}

# 下载二进制文件
download_binary() {
    BINARY_NAME="wecom-notifier-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi

    DOWNLOAD_URL="https://github.com/GhostLee/wecom-notifier/releases/download/${LATEST_VERSION}/${BINARY_NAME}.tar.gz"
    
    info "下载: $DOWNLOAD_URL"
    
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    if ! curl -fsSL -o "${BINARY_NAME}.tar.gz" "$DOWNLOAD_URL"; then
        error "下载失败"
    fi
    
    info "解压文件..."
    tar -xzf "${BINARY_NAME}.tar.gz"
    
    if [ ! -f "$BINARY_NAME" ]; then
        error "解压后未找到二进制文件"
    fi
}

# 安装到系统
install_binary() {
    INSTALL_DIR="/usr/local/bin"
    TARGET_NAME="wecom-notifier"
    
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="$HOME/bin"
        TARGET_NAME="wecom-notifier.exe"
    fi
    
    info "安装到: ${INSTALL_DIR}/${TARGET_NAME}"
    
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
    fi
    
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_NAME" "${INSTALL_DIR}/${TARGET_NAME}"
        chmod +x "${INSTALL_DIR}/${TARGET_NAME}"
    else
        info "需要 sudo 权限安装到系统目录"
        sudo mv "$BINARY_NAME" "${INSTALL_DIR}/${TARGET_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${TARGET_NAME}"
    fi
    
    info "安装完成！"
}

# 创建配置文件
create_config() {
    CONFIG_DIR="$HOME/.config/wecom-notifier"
    mkdir -p "$CONFIG_DIR"
    
    if [ ! -f "${CONFIG_DIR}/.env" ]; then
        info "创建配置文件: ${CONFIG_DIR}/.env"
        cat > "${CONFIG_DIR}/.env" << EOF
# 企业微信配置
WECOM_CORP_ID=your_corp_id
WECOM_AGENT_ID=1000002
WECOM_SECRET=your_secret
WECOM_TO_USER=@all

# API 认证
API_KEY=your_secure_api_key_here

# 服务器配置
PORT=8080
GIN_MODE=release

# 日志配置
LOG_LEVEL=info
LOG_DIR=${CONFIG_DIR}/logs
LOG_MAX_AGE_DAYS=30
LOG_ROTATE=true
EOF
        warn "请编辑配置文件: ${CONFIG_DIR}/.env"
    else
        info "配置文件已存在，跳过创建"
    fi
}

# 安装 systemd 服务（仅 Linux）
install_systemd_service() {
    if [ "$OS" != "linux" ]; then
        return
    fi
    
    read -p "是否安装为 systemd 服务？ [y/N] " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "创建 systemd 服务..."
        
        SERVICE_FILE="/etc/systemd/system/wecom-notifier.service"
        
        sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=WeCom Notifier Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/.config/wecom-notifier
EnvironmentFile=$HOME/.config/wecom-notifier/.env
ExecStart=/usr/local/bin/wecom-notifier
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
        
        sudo systemctl daemon-reload
        sudo systemctl enable wecom-notifier
        
        info "systemd 服务已安装"
        info "启动服务: sudo systemctl start wecom-notifier"
        info "查看状态: sudo systemctl status wecom-notifier"
        info "查看日志: journalctl -u wecom-notifier -f"
    fi
}

# 清理临时文件
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

# 主函数
main() {
    info "开始安装企业微信通知服务..."
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    create_config
    install_systemd_service
    cleanup
    
    echo
    info "======================================"
    info "安装成功！"
    info "======================================"
    echo
    info "配置文件: $HOME/.config/wecom-notifier/.env"
    info "日志目录: $HOME/.config/wecom-notifier/logs"
    echo
    info "运行命令: wecom-notifier"
    info "测试页面: http://localhost:8080"
    echo
    warn "请先编辑配置文件再启动服务！"
}

# 设置清理钩子
trap cleanup EXIT

# 运行主函数
main