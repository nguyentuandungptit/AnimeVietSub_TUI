package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	HistoryFilePath = ".avs_history.json"
	MaxHistoryItems = 50
)

// HistoryItem lưu lại một tập anime đã xem.
type HistoryItem struct {
	AnimeTitle   string    `json:"anime_title"`
	AnimeURL     string    `json:"anime_url"`
	Poster       string    `json:"poster"`
	EpisodeTitle string    `json:"episode_title"`
	EpisodeID    string    `json:"episode_id"`
	EpisodeHash  string    `json:"episode_hash"`
	WatchedAt    time.Time `json:"watched_at"`
}

var histMu sync.Mutex

// LoadHistory nạp danh sách lịch sử xem phim từ file.
func LoadHistory() []HistoryItem {
	histMu.Lock()
	defer histMu.Unlock()

	data, err := os.ReadFile(HistoryFilePath)
	if err != nil {
		return []HistoryItem{}
	}

	var items []HistoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		return []HistoryItem{}
	}

	return items
}

// AddHistory thêm hoặc cập nhật một mục vào lịch sử xem và đưa lên đầu danh sách qua atomic write.
func AddHistory(item HistoryItem) error {
	histMu.Lock()
	defer histMu.Unlock()

	var items []HistoryItem
	data, err := os.ReadFile(HistoryFilePath)
	if err == nil {
		_ = json.Unmarshal(data, &items)
	}

	item.WatchedAt = time.Now()

	// Lọc bỏ mục cũ cùng AnimeURL nếu đã tồn tại
	var newItems []HistoryItem
	newItems = append(newItems, item)

	for _, it := range items {
		if it.AnimeURL != item.AnimeURL {
			newItems = append(newItems, it)
		}
		if len(newItems) >= MaxHistoryItems {
			break
		}
	}

	out, err := json.MarshalIndent(newItems, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write
	dir := filepath.Dir(HistoryFilePath)
	tmpFile, err := os.CreateTemp(dir, "avs_history_*.tmp")
	if err != nil {
		return os.WriteFile(HistoryFilePath, out, 0644)
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(out); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("ghi file tạm lịch sử thất bại: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("sync file tạm lịch sử thất bại: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("đóng file tạm lịch sử thất bại: %w", err)
	}

	if err := os.Rename(tmpName, HistoryFilePath); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("đổi tên file lịch sử thất bại: %w", err)
	}

	return nil
}

// ClearHistory xóa sạch lịch sử xem phim.
func ClearHistory() error {
	histMu.Lock()
	defer histMu.Unlock()

	_ = os.Remove(HistoryFilePath)
	return nil
}
