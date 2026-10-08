#!/usr/bin/env bash

set -e

# Mã màu giao diện
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

REPO="nguyentuandungptit/AnimeVietSub_TUI"
BINARY_NAME="avs"

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

echo -e "${BOLD}${BLUE}"
echo -e "${NC}"
echo -e "${BOLD}Đang cài đặt AnimeVietSub TUI (avs)...${NC}\n"

# 1. Xác định Hệ điều hành và Kiến trúc
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        TARGET_ARCH="amd64"
        ;;
    aarch64|arm64)
        TARGET_ARCH="arm64"
        ;;
    armv7l|armv6l)
        TARGET_ARCH="arm"
        ;;
    *)
        error "Kiến trúc CPU không được hỗ trợ: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux)
        TARGET_OS="linux"
        ;;
    darwin)
        TARGET_OS="darwin"
        ;;
    msys*|cygwin*|mingw*)
        TARGET_OS="windows"
        BINARY_NAME="avs.exe"
        ;;
    *)
        error "Hệ điều hành không được hỗ trợ: $OS"
        exit 1
        ;;
esac

info "Hệ thống: $TARGET_OS ($TARGET_ARCH)"

# 2. Xác định thư mục cài đặt mục tiêu
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
    SUDO=""
elif command -v sudo >/dev/null 2>&1 && [ "$EUID" -ne 0 ]; then
    # Hỏi quyền sudo hoặc fallback nếu người dùng từ chối
    INSTALL_DIR="/usr/local/bin"
    SUDO="sudo"
else
    INSTALL_DIR="$HOME/.local/bin"
    SUDO=""
fi

# Cho phép ghi đè đường dẫn qua biến môi trường INSTALL_DIR
if [ -n "$CUSTOM_INSTALL_DIR" ]; then
    INSTALL_DIR="$CUSTOM_INSTALL_DIR"
    SUDO=""
fi

info "Thư mục cài đặt: $INSTALL_DIR"
$SUDO mkdir -p "$INSTALL_DIR"

# 3. Tìm hoặc tải binary
TMP_DIR=$(mktemp -d)
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

INSTALLED_SUCCESS=0

# Cách 3a: Nếu đang chạy trực tiếp từ thư mục chứa mã nguồn/build sẵn
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" >/dev/null 2>&1 && pwd)"
if [ -f "$SCRIPT_DIR/$BINARY_NAME" ]; then
    info "Phát hiện binary có sẵn tại thư mục: $SCRIPT_DIR/$BINARY_NAME"
    $SUDO cp -f "$SCRIPT_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    $SUDO chmod 755 "$INSTALL_DIR/$BINARY_NAME"
    INSTALLED_SUCCESS=1
elif [ -f "./$BINARY_NAME" ]; then
    info "Phát hiện binary có sẵn trong thư mục hiện tại."
    $SUDO cp -f "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    $SUDO chmod 755 "$INSTALL_DIR/$BINARY_NAME"
    INSTALLED_SUCCESS=1
fi

# Cách 3b: Tải từ GitHub Releases (nếu chưa cài từ local)
if [ $INSTALLED_SUCCESS -eq 0 ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}_${TARGET_OS}_${TARGET_ARCH}"
    FALLBACK_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
    TARGET_BIN="$TMP_DIR/$BINARY_NAME"

    info "Đang thử tải bản phát hành từ GitHub Releases..."
    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL -L "$RELEASE_URL" -o "$TARGET_BIN" 2>/dev/null || curl -fsSL -L "$FALLBACK_URL" -o "$TARGET_BIN" 2>/dev/null; then
            $SUDO cp -f "$TARGET_BIN" "$INSTALL_DIR/$BINARY_NAME"
            $SUDO chmod 755 "$INSTALL_DIR/$BINARY_NAME"
            INSTALLED_SUCCESS=1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q -O "$TARGET_BIN" "$RELEASE_URL" 2>/dev/null || wget -q -O "$TARGET_BIN" "$FALLBACK_URL" 2>/dev/null; then
            $SUDO cp -f "$TARGET_BIN" "$INSTALL_DIR/$BINARY_NAME"
            $SUDO chmod 755 "$INSTALL_DIR/$BINARY_NAME"
            INSTALLED_SUCCESS=1
        fi
    fi
fi

# Cách 3c: Nếu không tải được release, kiểm tra Go để tự build tự động
if [ $INSTALLED_SUCCESS -eq 0 ]; then
    if command -v go >/dev/null 2>&1; then
        info "Không tìm thấy pre-built release, hệ thống có Go -> Đang tự động biên dịch mã nguồn..."
        BUILD_SRC_DIR="$TMP_DIR/src"
        if [ -d "$SCRIPT_DIR/cmd/avs" ]; then
            BUILD_SRC_DIR="$SCRIPT_DIR"
        else
            info "Đang clone mã nguồn mới nhất từ GitHub..."
            git clone --depth 1 "https://github.com/${REPO}.git" "$BUILD_SRC_DIR"
        fi
        (
            cd "$BUILD_SRC_DIR"
            go build -ldflags="-s -w" -o "$TMP_DIR/$BINARY_NAME" cmd/avs/main.go
        )
        $SUDO cp -f "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        $SUDO chmod 755 "$INSTALL_DIR/$BINARY_NAME"
        INSTALLED_SUCCESS=1
    else
        error "Không thể tải binary build sẵn và máy bạn chưa cài đặt Golang để tự build."
        echo -e "Vui lòng cài đặt Go (https://go.dev) hoặc tải binary thủ công từ:"
        echo -e "https://github.com/${REPO}/releases\n"
        exit 1
    fi
fi

if [ $INSTALLED_SUCCESS -eq 1 ]; then
    success "Đã cài đặt thành công '$BINARY_NAME' vào: $INSTALL_DIR/$BINARY_NAME"
fi

# 4. Kiểm tra biến môi trường PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    warn "Thư mục '$INSTALL_DIR' hiện chưa nằm trong biến môi trường PATH của bạn."
    echo -e "Để chạy lệnh 'avs' từ mọi nơi, hãy thêm dòng sau vào ~/.bashrc hoặc ~/.zshrc:"
    echo -e "    ${BOLD}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}\n"
fi

# 5. Kiểm tra và hướng dẫn cài đặt MPV (Trình phát video khuyến nghị)
echo ""
info "Kiểm tra trình phát video (mpv)..."
if command -v mpv >/dev/null 2>&1; then
    success "mpv đã được cài đặt ($(mpv --version | head -n 1))."
else
    warn "mpv chưa được cài đặt trên máy của bạn!"
    echo -e "AnimeVietSub TUI sử dụng mpv để phát video mượt mà và trực tiếp từ terminal."
    echo -e "Bạn có thể cài đặt nhanh mpv bằng lệnh sau:"
    case "$TARGET_OS" in
        linux)
            if command -v apt >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo apt update && sudo apt install -y mpv${NC}"
            elif command -v dnf >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo dnf install -y mpv${NC}"
            elif command -v pacman >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo pacman -S mpv${NC}"
            elif command -v zypper >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo zypper install mpv${NC}"
            else
                echo -e "    Cài đặt mpv qua trình quản lý gói của bản phân phối Linux đang dùng."
            fi
            ;;
        darwin)
            echo -e "    ${BOLD}brew install mpv${NC}"
            ;;
        windows)
            echo -e "    ${BOLD}winget install shinchiro.mpv${NC}"
            ;;
    esac
    echo -e "(Lưu ý: Ứng dụng vẫn có thể mở phim bằng trình duyệt web nếu chưa có mpv)\n"
fi

echo -e "${BOLD}${GREEN}================================================================${NC}"
echo -e "${BOLD}${GREEN}  CÀI ĐẶT HOÀN TẤT! Chúc bạn xem anime vui vẻ! (｡♥‿♥｡)${NC}"
echo -e "${BOLD}${GREEN}================================================================${NC}"
echo -e "Chạy lệnh sau để bắt đầu thưởng thức:"
echo -e "    ${BOLD}${CYAN}avs${NC}\n"
