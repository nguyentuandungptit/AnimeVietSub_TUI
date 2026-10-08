#!/usr/bin/env bash
# ==============================================================================
#  AnimeVietSub TUI (avs) - Trình Gỡ Cài Đặt Tự Động (Uninstaller)
#  GitHub: https://github.com/nguyentuandungptit/AnimeVietSub_TUI
# ==============================================================================

set -e

# Mã màu giao diện
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

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

echo -e "${BOLD}${RED}"
cat << 'EOF'
     ___         _               __   ___     _   ___       _   
    / _ \       (_)             \ \ / (_)   | | / ___|     | |  
   / /_\ \_ __   _ _ __ ___   ___\ V / _ ___| |_\ `--. _   _| |__  
   |  _  | '_ \ | | '_ ` _ \ / _ \   / | / __| '_ `--. \ | | | '_ \ 
   | | | | | | || | | | | | |  __/ | |  | \__ \ |_) /\__/ / |_| | |_) |
   \_| |_/_| |_|/ |_| |_| |_|\___\_/   |_|___/_.__/\____/ \__,_|_.__/  
EOF
echo -e "${NC}"
echo -e "${BOLD}Đang khởi chạy trình gỡ cài đặt AnimeVietSub TUI (avs)...${NC}\n"

# Kiểm tra cờ tham số
PURGE=0
for arg in "$@"; do
    case "$arg" in
        --purge|-p)
            PURGE=1
            ;;
        --help|-h)
            echo "Sử dụng: ./uninstall.sh [TÙY CHỌN]"
            echo "Tùy chọn:"
            echo "  --purge, -p    Xóa toàn bộ binary cùng tệp cấu hình và lịch sử xem phim"
            echo "  --help, -h     Hiển thị trợ giúp này"
            exit 0
            ;;
    esac
done

FOUND_BIN=0

# 1. Tìm các vị trí cài đặt của binary avs
LOCATIONS=(
    "/usr/local/bin/$BINARY_NAME"
    "$HOME/.local/bin/$BINARY_NAME"
    "/usr/bin/$BINARY_NAME"
    "/bin/$BINARY_NAME"
)

# Thêm đường dẫn từ command -v nếu có
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    RESOLVED_BIN="$(command -v "$BINARY_NAME")"
    LOCATIONS+=("$RESOLVED_BIN")
fi

# Loại bỏ trùng lặp danh sách
UNIQUE_LOCATIONS=($(printf "%s\n" "${LOCATIONS[@]}" | sort -u))

for bin_path in "${UNIQUE_LOCATIONS[@]}"; do
    if [ -f "$bin_path" ]; then
        FOUND_BIN=1
        info "Phát hiện binary tại: $bin_path"
        
        # Kiểm tra quyền ghi
        if [ -w "$bin_path" ] || [ -w "$(dirname "$bin_path")" ]; then
            rm -f "$bin_path"
            success "Đã xóa: $bin_path"
        elif command -v sudo >/dev/null 2>&1; then
            info "Cần quyền sudo để xóa $bin_path..."
            sudo rm -f "$bin_path"
            success "Đã xóa: $bin_path (bằng quyền sudo)"
        else
            warn "Không thể xóa $bin_path do thiếu quyền quản trị viên."
        fi
    fi
done

if [ $FOUND_BIN -eq 0 ]; then
    warn "Không tìm thấy binary '$BINARY_NAME' trong các đường dẫn tiêu chuẩn của hệ thống."
fi

# 2. Xóa cấu hình và lịch sử nếu có yêu cầu hoặc hỏi người dùng
CONFIG_FILES=(
    "$HOME/.avs_config.json"
    "$HOME/.avs_history.json"
    "./.avs_config.json"
    "./.avs_history.json"
)

# Nếu chưa bật PURGE và terminal có hỗ trợ tương tác TTY
if [ $PURGE -eq 0 ] && [ -t 0 ]; then
    echo ""
    read -p "Bạn có muốn xóa toàn bộ tệp cấu hình và lịch sử xem phim không? (y/N): " -r answer
    case "$answer" in
        [yY][eE][sS]|[yY])
            PURGE=1
            ;;
        *)
            PURGE=0
            ;;
    esac
fi

if [ $PURGE -eq 1 ]; then
    echo ""
    info "Đang dọn dẹp các tệp cấu hình và lịch sử dữ liệu..."
    for cfg in "${CONFIG_FILES[@]}"; do
        if [ -f "$cfg" ]; then
            rm -f "$cfg"
            success "Đã xóa: $cfg"
        fi
    done
fi

echo ""
echo -e "${BOLD}${GREEN}================================================================${NC}"
echo -e "${BOLD}${GREEN}  ĐÃ GỠ CÀI ĐẶT THÀNH CÔNG! Hẹn gặp lại bạn lần sau! (•‿•)${NC}"
echo -e "${BOLD}${GREEN}================================================================${NC}\n"
