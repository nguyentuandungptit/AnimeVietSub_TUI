package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ConfigFilePath là đường dẫn file cấu hình.
var ConfigFilePath = ".avs_config.json"

// Config chứa các thiết lập của ứng dụng.
type Config struct {
	Language                 string `json:"language"`                    // "vi" hoặc "en"
	AutoOpenBrowserForIframe bool   `json:"auto_open_browser_iframe"`    // Tự động mở browser khi gặp iframe/embed
	PlayerMode               string `json:"player_mode"`                 // "auto" (Standalone cho Iframe, MPV cho Direct), "app_window", "browser", "mpv"
	PreferredQuality         string `json:"preferred_quality"`           // "1080p", "720p", "auto"
	ThemeMode                string `json:"theme_mode"`                  // "terminal" (mặc định theo Kitty/Foot), "light", "dark"
}

var (
	defaultConfig = Config{
		Language:                 "vi",
		AutoOpenBrowserForIframe: false,
		PlayerMode:               "auto",
		PreferredQuality:         "1080p",
		ThemeMode:                "terminal",
	}
	cfgMu        sync.Mutex
	cachedConfig *Config
)

// LoadConfig nạp cấu hình từ RAM cache hoặc từ file nếu chưa có trong RAM.
func LoadConfig() *Config {
	cfgMu.Lock()
	defer cfgMu.Unlock()

	if cachedConfig != nil {
		copyCfg := *cachedConfig
		return &copyCfg
	}

	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		cfg := defaultConfig
		_ = saveConfigLocked(&cfg)
		cachedConfig = &cfg
		copyCfg := *cachedConfig
		return &copyCfg
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = defaultConfig
		cachedConfig = &cfg
		copyCfg := *cachedConfig
		return &copyCfg
	}

	// Đảm bảo các giá trị mặc định nếu rỗng
	if cfg.Language == "" {
		cfg.Language = "vi"
	}
	if cfg.PlayerMode == "" {
		cfg.PlayerMode = "auto"
	}
	if cfg.PreferredQuality == "" {
		cfg.PreferredQuality = "1080p"
	}
	if cfg.ThemeMode == "" {
		cfg.ThemeMode = "terminal"
	}

	cachedConfig = &cfg
	copyCfg := *cachedConfig
	return &copyCfg
}

// SaveConfig lưu cấu hình vào file qua cơ chế atomic write (ghi file tạm rồi đổi tên) và cập nhật cache RAM.
func SaveConfig(cfg *Config) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	return saveConfigLocked(cfg)
}

func saveConfigLocked(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: ghi ra file .tmp rồi rename
	dir := filepath.Dir(ConfigFilePath)
	tmpFile, err := os.CreateTemp(dir, "avs_config_*.tmp")
	if err != nil {
		// Fallback ghi thẳng nếu không tạo được temp file
		return os.WriteFile(ConfigFilePath, data, 0644)
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("ghi file tạm cấu hình thất bại: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("sync file cấu hình thất bại: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("đóng file tạm cấu hình thất bại: %w", err)
	}

	if err := os.Rename(tmpName, ConfigFilePath); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("đổi tên file cấu hình thất bại: %w", err)
	}

	copyCfg := *cfg
	cachedConfig = &copyCfg
	return nil
}
