# Makefile for AnimeVietSub TUI (avs)

BINARY_NAME := avs
MAIN_SRC    := cmd/avs/main.go
PREFIX      ?= /usr/local
INSTALL_BIN := $(DESTDIR)$(PREFIX)/bin

GO          := go
LDFLAGS     := -ldflags="-s -w"

.PHONY: all build run install uninstall clean tidy help

all: build

help:
	@echo "AnimeVietSub TUI (avs) - Build Commands:"
	@echo "  make build       - Biên dịch binary (mặc định: ./avs)"
	@echo "  make run         - Chạy trực tiếp từ mã nguồn"
	@echo "  make install     - Cài đặt binary vào $(INSTALL_BIN) (yêu cầu quyền sudo nếu cần)"
	@echo "  make uninstall   - Gỡ cài đặt binary khỏi $(INSTALL_BIN)"
	@echo "  make purge       - Gỡ cài đặt và xóa toàn bộ dữ liệu cấu hình, lịch sử"
	@echo "  make clean       - Xóa file binary đã biên dịch"
	@echo "  make tidy        - Dọn dẹp và cập nhật Go modules"

build:
	@echo "==> Đang biên dịch $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_SRC)
	@echo "==> Biên dịch thành công: ./$(BINARY_NAME)"

run:
	$(GO) run $(MAIN_SRC)

install: build
	@echo "==> Đang cài đặt $(BINARY_NAME) vào $(INSTALL_BIN)..."
	@mkdir -p $(INSTALL_BIN)
	@install -m 755 $(BINARY_NAME) $(INSTALL_BIN)/$(BINARY_NAME)
	@echo "==> Cài đặt thành công! Bạn có thể khởi chạy bằng lệnh: $(BINARY_NAME)"

uninstall:
	@echo "==> Đang gỡ bỏ $(BINARY_NAME) khỏi $(INSTALL_BIN)..."
	@rm -f $(INSTALL_BIN)/$(BINARY_NAME)
	@echo "==> Đã gỡ bỏ cài đặt thành công."

purge: uninstall
	@echo "==> Đang xóa tệp cấu hình và lịch sử xem..."
	@rm -f .avs_config.json .avs_history.json $(HOME)/.avs_config.json $(HOME)/.avs_history.json
	@echo "==> Đã dọn dẹp sạch sẽ toàn bộ dữ liệu."

clean:
	@echo "==> Đang dọn dẹp..."
	@rm -f $(BINARY_NAME)
	@echo "==> Hoàn tất."

tidy:
	$(GO) mod tidy
