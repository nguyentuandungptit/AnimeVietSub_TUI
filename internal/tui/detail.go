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

// SelectEpisodeMsg phát ra khi người dùng chọn một tập phim để resolve nguồn phát.
type SelectEpisodeMsg struct {
	Anime    *api.AnimeDetail
	AnimeURL string
	Episode  api.Episode
}

// detailLoadedMsg dữ liệu chi tiết phim tải về từ API.
type detailLoadedMsg struct {
	detail *api.AnimeDetail
	err    error
}

// DetailModel quản lý màn hình xem thông tin chi tiết phim và chọn tập phim.
type DetailModel struct {
	client    *api.Client
	animeItem api.AnimeItem
	detail    *api.AnimeDetail
	epCursor  int
	loading   bool
	err       error
	spinner   spinner.Model
	width     int
	height    int
}

// NewDetailModel khởi tạo DetailModel.
func NewDetailModel(client *api.Client) DetailModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return DetailModel{
		client:  client,
		spinner: s,
	}
}

// Init khởi động spinner.
func (m DetailModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// SetSize cập nhật kích thước khung nhìn.
func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// LoadAnime thiết lập anime cần xem và gửi request tải chi tiết.
func (m *DetailModel) LoadAnime(item api.AnimeItem) tea.Cmd {
	m.animeItem = item
	m.detail = nil
	m.epCursor = 0
	m.loading = true
	m.err = nil

	return tea.Batch(
		m.spinner.Tick,
		m.fetchDetail(item.URL),
	)
}

func (m DetailModel) fetchDetail(url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		detail, err := m.client.GetFilmDetail(ctx, url)
		return detailLoadedMsg{
			detail: detail,
			err:    err,
		}
	}
}

func (m DetailModel) calcCols() int {
	colWidth := 14
	cols := (m.width - 6) / colWidth
	if cols < 4 {
		cols = 4
	}
	if cols > 12 {
		cols = 12
	}
	return cols
}

// Update xử lý sự kiện trong màn hình chi tiết phim.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case detailLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.detail = msg.detail
		m.epCursor = 0
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.err != nil {
			if msg.String() == "r" {
				m.loading = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.fetchDetail(m.animeItem.URL))
			}
			return m, nil
		}

		if m.detail == nil || len(m.detail.Episodes) == 0 {
			return m, nil
		}

		totalEps := len(m.detail.Episodes)
		cols := m.calcCols()

		switch msg.String() {
		case "up", "k":
			if m.epCursor >= cols {
				m.epCursor -= cols
			}
			return m, nil

		case "down", "j":
			if m.epCursor+cols < totalEps {
				m.epCursor += cols
			} else {
				m.epCursor = totalEps - 1
			}
			return m, nil

		case "left", "h":
			if m.epCursor > 0 {
				m.epCursor--
			}
			return m, nil

		case "right", "l":
			if m.epCursor < totalEps-1 {
				m.epCursor++
			}
			return m, nil

		case "enter":
			if m.epCursor >= 0 && m.epCursor < totalEps {
				selectedEp := m.detail.Episodes[m.epCursor]
				animeURL := m.animeItem.URL
				if m.detail != nil && m.detail.URL != "" {
					animeURL = m.detail.URL
				}
				return m, func() tea.Msg {
					return SelectEpisodeMsg{
						Anime:    m.detail,
						AnimeURL: animeURL,
						Episode:  selectedEp,
					}
				}
			}

		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.fetchDetail(m.animeItem.URL))
		}
	}

	return m, nil
}

// View kết xuất màn hình chi tiết phim.
func (m DetailModel) View() string {
	var b strings.Builder

	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	cfg := config.LoadConfig()
	lang := cfg.Language

	if m.loading {
		title := m.animeItem.Title
		if title == "" {
			title = "anime"
		}
		b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "detail_loading", title)))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ErrorStyle.Render(i18n.T(lang, "detail_error", m.err)))
		return b.String()
	}

	if m.detail == nil {
		b.WriteString(i18n.T(lang, "detail_no_data"))
		return b.String()
	}

	// 1. Tiêu đề và Thông tin chung
	title := m.detail.Title
	if title == "" {
		title = m.animeItem.Title
	}
	b.WriteString(AppHeaderStyle.Render(TruncateDisplay("🎬 "+title, contentWidth)))
	b.WriteString("\n")

	var metaParts []string
	if m.detail.FilmID != "" {
		metaParts = append(metaParts, fmt.Sprintf("FilmID: %s", m.detail.FilmID))
	}
	if m.detail.Score != "" {
		metaParts = append(metaParts, fmt.Sprintf("★ %s", m.detail.Score))
	}
	metaParts = append(metaParts, i18n.T(lang, "detail_episodes", len(m.detail.Episodes)))
	b.WriteString(TabInactiveStyle.Render(strings.Join(metaParts, "  •  ")))
	b.WriteString("\n\n")

	// 2. Mô tả tóm tắt (cắt 2-3 dòng tối đa)
	if m.detail.Description != "" {
		descLines := strings.Split(m.detail.Description, "\n")
		cleanDesc := strings.TrimSpace(descLines[0])
		if len(descLines) > 1 && len(cleanDesc) < 80 {
			cleanDesc += " " + strings.TrimSpace(descLines[1])
		}
		truncatedDesc := TruncateDisplay(i18n.T(lang, "detail_desc", cleanDesc), contentWidth*2)
		b.WriteString(TabInactiveStyle.Render(truncatedDesc))
		b.WriteString("\n\n")
	}

	// 3. Danh sách tập phim dạng lưới
	b.WriteString(TabActiveStyle.Render(i18n.T(lang, "detail_ep_list")))
	b.WriteString("\n")

	if len(m.detail.Episodes) == 0 {
		b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "detail_ep_empty")))
		return b.String()
	}

	cols := m.calcCols()

	totalEps := len(m.detail.Episodes)
	maxGridRows := m.height - 8
	if maxGridRows < 3 {
		maxGridRows = 3
	}

	currentRow := m.epCursor / cols
	startRow := 0
	if currentRow >= maxGridRows {
		startRow = currentRow - maxGridRows + 1
	}
	endRow := startRow + maxGridRows

	for r := startRow; r < endRow; r++ {
		var rowItems []string
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= totalEps {
				break
			}
			ep := m.detail.Episodes[idx]
			label := ep.Title
			if label == "" {
				label = fmt.Sprintf("Tập %02d", idx+1)
			}
			// Rút gọn label ngắn gọn: "Tập 01" hoặc "01"
			cleanLabel := TruncateDisplay(label, 12)

			if idx == m.epCursor {
				rowItems = append(rowItems, SelectedItemStyle.Render(cleanLabel))
			} else {
				rowItems = append(rowItems, TabInactiveStyle.Render(cleanLabel))
			}
		}
		if len(rowItems) > 0 {
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
			b.WriteString("\n")
		}
	}

	return b.String()
}
