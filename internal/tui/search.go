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
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// searchSuggestionsMsg kết quả gợi ý tìm kiếm.
type searchSuggestionsMsg struct {
	reqID       int
	suggestions []api.SearchSuggestion
	err         error
}

// searchResultMsg kết quả tìm kiếm đầy đủ từ /tim-kiem/{kw}/.
type searchResultMsg struct {
	keyword string
	page    int
	result  *api.SearchResult
	err     error
}

// debounceSearchMsg thông báo hết thời gian chờ debounce 300ms.
type debounceSearchMsg struct {
	reqID   int
	keyword string
}

// SearchState trạng thái của màn hình tìm kiếm.
type SearchState int

const (
	SearchStateInput SearchState = iota
	SearchStateResults
)

// SearchModel quản lý ô nhập liệu tìm kiếm, gợi ý và kết quả tìm kiếm đầy đủ.
type SearchModel struct {
	client      *api.Client
	state       SearchState
	input       textinput.Model
	spinner     spinner.Model
	suggestions []api.SearchSuggestion
	suggCursor  int
	result      *api.SearchResult
	resCursor   int
	keyword     string
	currentPage int
	loading     bool
	err         error
	debounceID  int
	width       int
	height      int
}

// NewSearchModel khởi tạo SearchModel.
func NewSearchModel(client *api.Client) SearchModel {
	cfg := config.LoadConfig()
	ti := textinput.New()
	ti.Placeholder = i18n.T(cfg.Language, "search_placeholder")
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return SearchModel{
		client:      client,
		state:       SearchStateInput,
		input:       ti,
		spinner:     s,
		currentPage: 1,
	}
}

// Init kích hoạt focus và spinner.
func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

// SetSize cập nhật kích thước khung nhìn.
func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if w > 20 {
		m.input.Width = w - 10
	}
}

// Reset đặt lại ô nhập và kết quả tìm kiếm.
func (m *SearchModel) Reset() {
	m.state = SearchStateInput
	m.input.Reset()
	m.input.Focus()
	m.suggestions = nil
	m.suggCursor = 0
	m.result = nil
	m.resCursor = 0
	m.keyword = ""
	m.currentPage = 1
	m.err = nil
	m.loading = false
}

// triggerDebounce tạo lệnh chờ 300ms để debounce tìm kiếm gợi ý.
func (m *SearchModel) triggerDebounce(keyword string) tea.Cmd {
	m.debounceID++
	id := m.debounceID
	return tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
		return debounceSearchMsg{
			reqID:   id,
			keyword: keyword,
		}
	})
}

// fetchSuggestions gửi request POST /ajax/suggest để lấy gợi ý tìm kiếm.
func (m SearchModel) fetchSuggestions(reqID int, keyword string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		suggs, err := m.client.GetSearchSuggestions(ctx, keyword)
		return searchSuggestionsMsg{
			reqID:       reqID,
			suggestions: suggs,
			err:         err,
		}
	}
}

// fetchSearchResults gửi request tìm kiếm đầy đủ theo từ khóa và trang.
func (m SearchModel) fetchSearchResults(keyword string, page int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := m.client.SearchAnime(ctx, keyword, page)
		return searchResultMsg{
			keyword: keyword,
			page:    page,
			result:  res,
			err:     err,
		}
	}
}

// Update xử lý sự kiện trong màn hình tìm kiếm.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case debounceSearchMsg:
		if msg.reqID == m.debounceID && msg.keyword != "" && m.state == SearchStateInput {
			m.loading = true
			return m, m.fetchSuggestions(msg.reqID, msg.keyword)
		}

	case searchSuggestionsMsg:
		if msg.reqID == m.debounceID {
			m.loading = false
			m.suggestions = msg.suggestions
			m.suggCursor = -1
		}
		return m, nil

	case searchResultMsg:
		if msg.keyword == m.keyword {
			m.loading = false
			m.err = msg.err
			m.result = msg.result
			m.currentPage = msg.page
			m.resCursor = 0
		}
		return m, nil

	case tea.KeyMsg:
		// Khi đang ở trạng thái nhập từ khóa
		if m.state == SearchStateInput {
			switch msg.String() {
			case "enter":
				// Nếu người dùng đang chọn một gợi ý cụ thể trong danh sách
				if m.suggCursor >= 0 && m.suggCursor < len(m.suggestions) {
					s := m.suggestions[m.suggCursor]
					return m, func() tea.Msg {
						return SelectAnimeMsg{
							Item: api.AnimeItem{
								Title:  s.Title,
								URL:    s.URL,
								Poster: s.Poster,
							},
						}
					}
				}

				// Ngược lại, tìm kiếm toàn bộ kết quả theo từ khóa nhập vào
				val := strings.TrimSpace(m.input.Value())
				if val != "" {
					m.state = SearchStateResults
					m.keyword = val
					m.currentPage = 1
					m.loading = true
					m.err = nil
					m.input.Blur()
					return m, tea.Batch(m.spinner.Tick, m.fetchSearchResults(val, 1))
				}
				return m, nil

			case "up":
				if len(m.suggestions) > 0 {
					if m.suggCursor > 0 {
						m.suggCursor--
					} else if m.suggCursor == 0 {
						m.suggCursor = -1
					}
				}
				return m, nil

			case "down":
				if len(m.suggestions) > 0 && m.suggCursor < len(m.suggestions)-1 {
					m.suggCursor++
				}
				return m, nil

			default:
				var tiCmd tea.Cmd
				m.input, tiCmd = m.input.Update(msg)
				curVal := strings.TrimSpace(m.input.Value())
				m.suggCursor = -1
				if curVal != "" {
					return m, tea.Batch(tiCmd, m.triggerDebounce(curVal))
				} else {
					m.suggestions = nil
					m.loading = false
					return m, tiCmd
				}
			}
		}

		// Khi đang ở danh sách kết quả tìm kiếm
		if m.state == SearchStateResults {
			switch msg.String() {
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

			case "n", "right", "l": // Trang tiếp theo
				if m.result != nil && m.result.HasNext {
					m.loading = true
					m.err = nil
					return m, tea.Batch(m.spinner.Tick, m.fetchSearchResults(m.keyword, m.currentPage+1))
				}
				return m, nil

			case "p", "left", "h": // Trang trước
				if m.currentPage > 1 {
					m.loading = true
					m.err = nil
					return m, tea.Batch(m.spinner.Tick, m.fetchSearchResults(m.keyword, m.currentPage-1))
				}
				return m, nil

			case "enter":
				if m.result != nil && len(m.result.Items) > 0 && m.resCursor < len(m.result.Items) {
					item := m.result.Items[m.resCursor]
					return m, func() tea.Msg {
						return SelectAnimeMsg{Item: item}
					}
				}

			case "esc":
				// Quay lại thanh nhập từ khóa
				m.state = SearchStateInput
				m.input.Focus()
				return m, textinput.Blink

			case "r":
				m.loading = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.fetchSearchResults(m.keyword, m.currentPage))
			}
		}
	}

	return m, nil
}

// View kết xuất màn hình tìm kiếm.
func (m SearchModel) View() string {
	cfg := config.LoadConfig()
	lang := cfg.Language

	var b strings.Builder

	header := AppHeaderStyle.Render(i18n.T(lang, "search_header"))
	b.WriteString(header)
	b.WriteString("\n\n")

	// Ô nhập từ khóa
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "search_loading")))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(ErrorStyle.Render(i18n.T(lang, "search_error", m.err)))
		return b.String()
	}

	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	// 1. Giao diện Gợi ý tìm kiếm (khi ở SearchStateInput)
	if m.state == SearchStateInput {
		if len(m.suggestions) > 0 {
			b.WriteString(TabActiveStyle.Render(i18n.T(lang, "search_sugg_title")))
			b.WriteString("\n")
			for i, s := range m.suggestions {
				prefix := "  "
				if i == m.suggCursor {
					prefix = "> "
				}
				sub := ""
				if s.SubText != "" {
					sub = fmt.Sprintf("[%s]", s.SubText)
				}
				line := fmt.Sprintf("%s%-14s %s", prefix, sub, s.Title)
				if i == m.suggCursor {
					b.WriteString(SelectedItemStyle.Render(TruncateDisplay(line, contentWidth)))
				} else {
					b.WriteString(TruncateDisplay(line, contentWidth))
				}
				b.WriteString("\n")
			}
		}
		return b.String()
	}

	// 2. Giao diện Kết quả tìm kiếm đầy đủ (khi ở SearchStateResults)
	if m.result == nil || len(m.result.Items) == 0 {
		b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "search_empty")))
		return b.String()
	}

	// Thông tin phân trang
	pageInfo := i18n.T(lang, "search_page_info", m.keyword, m.currentPage, m.result.TotalPages)
	b.WriteString(TabActiveStyle.Render(pageInfo))
	b.WriteString("\n\n")

	maxListRows := m.height - 7
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

		availableForTitle := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(epsBadge) - lipgloss.Width(scoreText) - 4
		cleanTitle := TruncateDisplay(item.Title, availableForTitle)

		lineStr := fmt.Sprintf("%s%-12s %s", prefix, epsBadge, cleanTitle)
		if scoreText != "" {
			padding := contentWidth - lipgloss.Width(lineStr) - lipgloss.Width(scoreText)
			if padding > 0 {
				lineStr += strings.Repeat(" ", padding) + scoreText
			}
		}

		if i == m.resCursor {
			b.WriteString(SelectedItemStyle.Render(TruncateDisplay(lineStr, contentWidth)))
		} else {
			b.WriteString(TruncateDisplay(lineStr, contentWidth))
		}
		b.WriteString("\n")
	}

	return b.String()
}
