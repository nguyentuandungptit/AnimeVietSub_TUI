package tui

import (
	"fmt"
	"strings"
	"time"

	"avs-tui/internal/api"
	"avs-tui/internal/config"
	"avs-tui/internal/history"
	"avs-tui/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HistoryModel quản lý màn hình xem lại lịch sử các phim và tập đã xem.
type HistoryModel struct {
	items  []history.HistoryItem
	cursor int
	width  int
	height int
}

// NewHistoryModel khởi tạo HistoryModel.
func NewHistoryModel() HistoryModel {
	return HistoryModel{
		items:  history.LoadHistory(),
		cursor: 0,
	}
}

// Reload nạp lại lịch sử xem từ file.
func (m *HistoryModel) Reload() {
	m.items = history.LoadHistory()
	if m.cursor >= len(m.items) && len(m.items) > 0 {
		m.cursor = len(m.items) - 1
	}
}

// SetSize cập nhật kích thước khung nhìn.
func (m *HistoryModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update xử lý sự kiện trong màn hình lịch sử.
func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if len(m.items) > 0 && m.cursor < len(m.items) {
				it := m.items[m.cursor]
				return m, func() tea.Msg {
					return SelectAnimeMsg{
						Item: api.AnimeItem{
							Title: it.AnimeTitle,
							URL:   it.AnimeURL,
						},
					}
				}
			}

		case "d": // Xóa toàn bộ lịch sử
			_ = history.ClearHistory()
			m.items = nil
			m.cursor = 0
			return m, nil

		case "r":
			m.Reload()
			return m, nil
		}
	}

	return m, nil
}

// formatTimeAgo định dạng thời gian thân thiện.
func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "vừa xong"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d phút trước", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d giờ trước", int(diff.Hours()))
	}
	return t.Format("02/01/2006")
}

// View kết xuất giao diện lịch sử xem.
func (m HistoryModel) View() string {
	cfg := config.LoadConfig()
	lang := cfg.Language

	var b strings.Builder

	header := AppHeaderStyle.Render(i18n.T(lang, "history_header"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "history_empty")))
		return b.String()
	}

	subHeader := i18n.T(lang, "history_sub_header", len(m.items))
	b.WriteString(TabActiveStyle.Render(subHeader))
	b.WriteString("\n\n")

	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	maxListRows := m.height - 5
	if maxListRows < 5 {
		maxListRows = 5
	}

	startIdx := 0
	if m.cursor >= maxListRows {
		startIdx = m.cursor - maxListRows + 1
	}
	endIdx := startIdx + maxListRows
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.items[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		epBadge := fmt.Sprintf("[%s]", item.EpisodeTitle)
		timeStr := i18n.FormatTimeAgo(lang, item.WatchedAt)

		badgeFormatted := fmt.Sprintf("%-14s ", epBadge)
		availableForTitle := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(badgeFormatted) - lipgloss.Width(timeStr) - 2
		if availableForTitle < 10 {
			availableForTitle = 10
		}
		cleanTitle := TruncateDisplay(item.AnimeTitle, availableForTitle)

		lineLeft := fmt.Sprintf("%s%s%s", prefix, badgeFormatted, cleanTitle)
		padding := contentWidth - lipgloss.Width(lineLeft) - lipgloss.Width(timeStr)
		if padding < 1 {
			padding = 1
		}
		lineStr := lineLeft + strings.Repeat(" ", padding) + timeStr
		if lipgloss.Width(lineStr) > contentWidth {
			lineStr = TruncateDisplay(lineStr, contentWidth)
		}

		if i == m.cursor {
			b.WriteString(SelectedItemStyle.Render(lineStr))
		} else {
			b.WriteString(lineStr)
		}
		b.WriteString("\n")
	}

	return b.String()
}
