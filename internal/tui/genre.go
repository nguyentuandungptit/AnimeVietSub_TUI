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

type genreResultMsg struct {
	slug   string
	page   int
	result *api.SearchResult
	err    error
}

// GenreModel quản lý màn hình duyệt phim theo thể loại.
type GenreModel struct {
	client      *api.Client
	genres      []api.GenreItem
	genreCursor int
	result      *api.SearchResult
	resCursor   int
	currentPage int
	loading     bool
	err         error
	spinner     spinner.Model
	width       int
	height      int
}

// NewGenreModel khởi tạo GenreModel.
func NewGenreModel(client *api.Client) GenreModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return GenreModel{
		client:      client,
		genres:      api.DefaultGenres,
		genreCursor: 0,
		currentPage: 1,
		spinner:     s,
		loading:     true,
	}
}

// Init khởi động tải thể loại đầu tiên.
func (m GenreModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadCurrentGenre(),
	)
}

// SetSize cập nhật kích thước khung nhìn.
func (m *GenreModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m GenreModel) loadCurrentGenre() tea.Cmd {
	slug := m.genres[m.genreCursor].Slug
	page := m.currentPage
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := m.client.GetGenreItems(ctx, slug, page)
		return genreResultMsg{
			slug:   slug,
			page:   page,
			result: res,
			err:    err,
		}
	}
}

// Update xử lý sự kiện trong màn hình duyệt thể loại.
func (m GenreModel) Update(msg tea.Msg) (GenreModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case genreResultMsg:
		if msg.slug == m.genres[m.genreCursor].Slug {
			m.loading = false
			m.err = msg.err
			m.result = msg.result
			m.currentPage = msg.page
			m.resCursor = 0
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right", "l":
			if m.genreCursor < len(m.genres)-1 {
				m.genreCursor++
			} else {
				m.genreCursor = 0
			}
			m.currentPage = 1
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentGenre())

		case "shift+tab", "left", "h":
			if m.genreCursor > 0 {
				m.genreCursor--
			} else {
				m.genreCursor = len(m.genres) - 1
			}
			m.currentPage = 1
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentGenre())

		case "up", "k":
			if m.resCursor > 0 {
				m.resCursor--
			}
			return m, nil

		case "down", "j":
			if m.result != nil && m.resCursor < len(m.result.Items)-1 {
				m.resCursor++
			}
			return m, nil

		case "n": // Trang sau
			if m.result != nil && m.result.HasNext {
				m.currentPage++
				m.loading = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.loadCurrentGenre())
			}

		case "p": // Trang trước
			if m.currentPage > 1 {
				m.currentPage--
				m.loading = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.loadCurrentGenre())
			}

		case "enter":
			if m.result != nil && len(m.result.Items) > 0 && m.resCursor < len(m.result.Items) {
				selected := m.result.Items[m.resCursor]
				return m, func() tea.Msg {
					return SelectAnimeMsg{Item: selected}
				}
			}

		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.loadCurrentGenre())
		}
	}

	return m, nil
}

// View kết xuất màn hình thể loại.
func (m GenreModel) View() string {
	var b strings.Builder

	// 1. Render danh sách thể loại thanh trên bằng sliding window quanh genreCursor
	maxTabWidth := m.width - 6
	if maxTabWidth < 20 {
		maxTabWidth = 20
	}

	totalGenres := len(m.genres)
	renderedTabs := make([]string, totalGenres)
	widths := make([]int, totalGenres)
	for i, g := range m.genres {
		if i == m.genreCursor {
			renderedTabs[i] = TabActiveStyle.Render(g.Name)
		} else {
			renderedTabs[i] = TabInactiveStyle.Render(g.Name)
		}
		widths[i] = lipgloss.Width(renderedTabs[i])
	}

	start := m.genreCursor
	end := m.genreCursor
	curWidth := widths[m.genreCursor]

	for {
		expanded := false
		if end+1 < totalGenres {
			extra := widths[end+1]
			indicator := 0
			if end+1 < totalGenres-1 {
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
		if start-1 >= 0 {
			extra := widths[start-1]
			indicator := 0
			if start-1 > 0 {
				indicator += 4
			}
			if end < totalGenres-1 {
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
	if end < totalGenres-1 {
		tabViews = append(tabViews, TabInactiveStyle.Render("»"))
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
	b.WriteString(tabsRow)
	b.WriteString("\n\n")

	cfg := config.LoadConfig()
	lang := cfg.Language

	// 2. Nội dung danh sách anime
	curName := m.genres[m.genreCursor].Name
	if m.loading {
		b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "genre_loading", curName)))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ErrorStyle.Render(i18n.T(lang, "genre_error", m.err)))
		return b.String()
	}

	if m.result == nil || len(m.result.Items) == 0 {
		b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "genre_empty")))
		return b.String()
	}

	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Thông tin trang
	pageInfo := i18n.T(lang, "genre_page_info", curName, m.currentPage, m.result.TotalPages)
	b.WriteString(TabActiveStyle.Render(pageInfo))
	b.WriteString("\n\n")

	maxListRows := m.height - 5
	if maxListRows < 5 {
		maxListRows = 5
	}

	startIdx := 0
	if m.resCursor >= maxListRows {
		startIdx = m.resCursor - maxListRows + 1
	}
	endIdx := startIdx + maxListRows
	if endIdx > len(m.result.Items) {
		endIdx = len(m.result.Items)
	}

	for i := startIdx; i < endIdx; i++ {
		item := m.result.Items[i]
		prefix := "  "
		if i == m.resCursor {
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

		if i == m.resCursor {
			b.WriteString(SelectedItemStyle.Render(lineStr))
		} else {
			b.WriteString(lineStr)
		}
		b.WriteString("\n")
	}

	return b.String()
}
