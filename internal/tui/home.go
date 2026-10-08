package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"avs-tui/internal/api"
	"avs-tui/internal/config"
	"avs-tui/internal/i18n"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectAnimeMsg thông báo người dùng đã chọn một anime để xem chi tiết.
type SelectAnimeMsg struct {
	Item api.AnimeItem
}

// homeItemsLoadedMsg chứa dữ liệu anime tải về từ API.
type homeItemsLoadedMsg struct {
	tabIndex int
	items    []api.AnimeItem
	err      error
}

// HomeModel quản lý màn hình trang chủ với 11 tabs lọc phim.
type HomeModel struct {
	client    *api.Client
	tabs      []api.HomeTab
	activeTab int
	items     []api.AnimeItem
	cursor    int
	loading   bool
	err       error
	spinner   spinner.Model
	width     int
	height    int
}

// NewHomeModel khởi tạo HomeModel với spinner và danh sách tab.
func NewHomeModel(client *api.Client) HomeModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return HomeModel{
		client:    client,
		tabs:      api.HomeTabs,
		activeTab: 0,
		spinner:   s,
		loading:   true,
	}
}

// Init kích hoạt tải tab đầu tiên và khởi động spinner.
func (m HomeModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadCurrentTab(),
	)
}

// SetSize cập nhật kích thước khung nhìn cho HomeModel.
func (m *HomeModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// loadCurrentTab tạo tea.Cmd để tải dữ liệu tab hiện tại qua /ajax/item.
func (m HomeModel) loadCurrentTab() tea.Cmd {
	tab := m.tabs[m.activeTab]
	tabIdx := m.activeTab
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		items, err := m.client.GetHomeItems(ctx, tab.Type)
		return homeItemsLoadedMsg{
			tabIndex: tabIdx,
			items:    items,
			err:      err,
		}
	}
}

// Update xử lý sự kiện phím và dữ liệu trả về cho HomeModel.
func (m HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case homeItemsLoadedMsg:
		if msg.tabIndex == m.activeTab {
			m.loading = false
			m.err = msg.err
			m.items = msg.items
			m.cursor = 0
		}
		return m, nil

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

		case "left", "h":
			if m.activeTab > 0 {
				m.activeTab--
			} else {
				m.activeTab = len(m.tabs) - 1
			}
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentTab())

		case "right", "l", "tab":
			if m.activeTab < len(m.tabs)-1 {
				m.activeTab++
			} else {
				m.activeTab = 0
			}
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentTab())

		case "shift+tab":
			if m.activeTab > 0 {
				m.activeTab--
			} else {
				m.activeTab = len(m.tabs) - 1
			}
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentTab())

		case "enter":
			if len(m.items) > 0 && m.cursor < len(m.items) {
				selected := m.items[m.cursor]
				return m, func() tea.Msg {
					return SelectAnimeMsg{Item: selected}
				}
			}

		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentTab())
		}
	}

	return m, nil
}

// View kết xuất giao diện trang chủ với thanh tab và danh sách anime.
func (m HomeModel) View() string {
	cfg := config.LoadConfig()
	lang := cfg.Language

	var b strings.Builder

	// 1. Render tabs thanh trên bằng sliding window quanh activeTab
	maxTabWidth := m.width - 6
	if maxTabWidth < 20 {
		maxTabWidth = 20
	}

	totalTabs := len(m.tabs)
	renderedTabs := make([]string, totalTabs)
	widths := make([]int, totalTabs)
	for i, tab := range m.tabs {
		label := i18n.TranslateTab(lang, tab.Type, tab.Label)
		if i == m.activeTab {
			renderedTabs[i] = TabActiveStyle.Render(label)
		} else {
			renderedTabs[i] = TabInactiveStyle.Render(label)
		}
		widths[i] = lipgloss.Width(renderedTabs[i])
	}

	start := m.activeTab
	end := m.activeTab
	curWidth := widths[m.activeTab]

	for {
		expanded := false
		// Thử mở rộng sang phải
		if end+1 < totalTabs {
			extra := widths[end+1]
			indicator := 0
			if end+1 < totalTabs-1 {
				indicator += 4
			}
			if start > 0 {
				indicator += 4
			}
			if curWidth+extra+indicator <= maxTabWidth {
				end++
				curWidth += extra
				expanded = true
			}
		}
		// Thử mở rộng sang trái
		if start-1 >= 0 {
			extra := widths[start-1]
			indicator := 0
			if start-1 > 0 {
				indicator += 4
			}
			if end < totalTabs-1 {
				indicator += 4
			}
			if curWidth+extra+indicator <= maxTabWidth {
				start--
				curWidth += extra
				expanded = true
			}
		}
		if !expanded {
			break
		}
	}

	var tabViews []string
	if start > 0 {
		tabViews = append(tabViews, TabInactiveStyle.Render("«"))
	}
	for i := start; i <= end; i++ {
		tabViews = append(tabViews, renderedTabs[i])
	}
	if end < totalTabs-1 {
		tabViews = append(tabViews, TabInactiveStyle.Render("»"))
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
	b.WriteString(tabsRow)
	b.WriteString("\n\n")

	// 2. Nội dung danh sách anime
	activeLabel := i18n.TranslateTab(lang, m.tabs[m.activeTab].Type, m.tabs[m.activeTab].Label)
	if m.loading {
		loadingText := fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "home_loading", activeLabel))
		b.WriteString(loadingText)
		return b.String()
	}

	if m.err != nil {
		errBox := ErrorStyle.Render(i18n.T(lang, "home_error", m.err))
		b.WriteString(errBox)
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "home_empty")))
		return b.String()
	}

	// Tính toán số dòng hiển thị dựa trên chiều cao terminal
	maxListRows := m.height - 3
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

	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.items[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		epsBadge := ""
		if item.EpisodeStatus != "" {
			epsBadge = fmt.Sprintf("[%s]", item.EpisodeStatus)
		}

		scoreText := ""
		if item.Score != "" {
			scoreText = fmt.Sprintf("★ %s", item.Score)
		}

		badgeFormatted := ""
		if epsBadge != "" {
			badgeFormatted = fmt.Sprintf("%-12s ", epsBadge)
		}

		availableForTitle := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(badgeFormatted) - lipgloss.Width(scoreText) - 2
		if availableForTitle < 10 {
			availableForTitle = 10
		}
		cleanTitle := TruncateDisplay(item.Title, availableForTitle)

		lineLeft := fmt.Sprintf("%s%s%s", prefix, badgeFormatted, cleanTitle)
		padding := contentWidth - lipgloss.Width(lineLeft) - lipgloss.Width(scoreText)
		if padding < 1 {
			padding = 1
		}
		lineStr := lineLeft + strings.Repeat(" ", padding) + scoreText
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
