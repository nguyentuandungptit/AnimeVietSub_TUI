package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"avs-tui/internal/api"
	"avs-tui/internal/config"
	"avs-tui/internal/i18n"
	"avs-tui/internal/player"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen định danh màn hình hiện tại.
type Screen int

const (
	ScreenHome Screen = iota
	ScreenSearch
	ScreenDetail
	ScreenServers
	ScreenGenre
	ScreenHistory
	ScreenConfig
)

// AppModel root model điều hướng giữa các màn hình và quản lý vòng đời ứng dụng.
type AppModel struct {
	client        *api.Client
	currentScreen Screen
	screenStack   []Screen
	home          HomeModel
	search        SearchModel
	detail        DetailModel
	servers       ServersModel
	genre         GenreModel
	history       HistoryModel
	config        ConfigModel

	// Thanh tìm kiếm trên cùng (Top Bar)
	topSearch     textinput.Model
	searchFocused bool

	keys      KeyMap
	help      help.Model
	showHelp  bool
	width     int
	height    int
	statusMsg string
	cfg       config.Config
}

// NewAppModel khởi tạo AppModel với đầy đủ sub-models.
func NewAppModel(client *api.Client) AppModel {
	cfg := config.LoadConfig()
	ApplyTheme(cfg.ThemeMode)

	h := help.New()
	h.ShowAll = false

	ti := textinput.New()
	ti.Placeholder = i18n.T(cfg.Language, "top_search_hint")
	ti.CharLimit = 100
	ti.Width = 40

	return AppModel{
		client:        client,
		currentScreen: ScreenHome,
		home:          NewHomeModel(client),
		search:        NewSearchModel(client),
		detail:        NewDetailModel(client),
		servers:       NewServersModel(client),
		genre:         NewGenreModel(client),
		history:       NewHistoryModel(),
		config:        NewConfigModel(),
		topSearch:     ti,
		searchFocused: false,
		keys:          NewKeyMap(cfg.Language),
		help:          h,
		showHelp:      false,
		width:         80,
		height:        24,
		cfg:           *cfg,
	}
}

// Init khởi chạy ứng dụng, warmup cookies và tải dữ liệu trang chủ.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.warmupClient(),
		m.home.Init(),
		m.genre.Init(),
		textinput.Blink,
	)
}

func (m AppModel) warmupClient() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = m.client.Warmup(ctx)
		return nil
	}
}

// navigateTo chuyển sang màn hình mới và lưu lịch sử vào stack.
func (m *AppModel) navigateTo(s Screen) {
	if m.currentScreen != s {
		m.screenStack = append(m.screenStack, m.currentScreen)
		m.currentScreen = s
	}
}

// navigateBack quay lại màn hình trước đó trong lịch sử.
func (m *AppModel) navigateBack() {
	if len(m.screenStack) > 0 {
		prev := m.screenStack[len(m.screenStack)-1]
		m.screenStack = m.screenStack[:len(m.screenStack)-1]
		m.currentScreen = prev
	} else if m.currentScreen != ScreenHome {
		m.currentScreen = ScreenHome
	}
}

func (m *AppModel) syncConfig() {
	cfg := config.LoadConfig()
	m.cfg = *cfg
	m.keys = NewKeyMap(cfg.Language)
	m.topSearch.Placeholder = i18n.T(cfg.Language, "top_search_hint")
	ApplyTheme(cfg.ThemeMode)
}

// Update xử lý vòng lặp Model/Update/View của Root Model.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	lang := m.cfg.Language

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		panelInnerHeight := m.height - 5
		if panelInnerHeight < 10 {
			panelInnerHeight = 10
		}

		m.home.SetSize(m.width, panelInnerHeight)
		m.search.SetSize(m.width, panelInnerHeight)
		m.detail.SetSize(m.width, panelInnerHeight)
		m.servers.SetSize(m.width, panelInnerHeight)
		m.genre.SetSize(m.width, panelInnerHeight)
		m.history.SetSize(m.width, panelInnerHeight)
		m.config.SetSize(m.width, panelInnerHeight)

		searchWidth := m.width/2 - 10
		if searchWidth < 25 {
			searchWidth = 25
		}
		m.topSearch.Width = searchWidth
		return m, nil

	case SelectAnimeMsg:
		m.statusMsg = ""
		m.navigateTo(ScreenDetail)
		cmd := m.detail.LoadAnime(msg.Item)
		return m, cmd

	case SelectEpisodeMsg:
		m.statusMsg = ""
		m.navigateTo(ScreenServers)
		cmd := m.servers.LoadServers(msg.Anime.Title, msg.AnimeURL, msg.Episode)
		return m, cmd

	case PlayVideoMsg:
		if !player.IsMPVInstalled() {
			m.statusMsg = i18n.T(lang, "mpv_not_installed")
			return m, nil
		}
		m.statusMsg = i18n.T(lang, "playing_title", msg.Title)
		return m, player.PlayStream(msg.VideoURL, msg.Title)

	case player.FinishedMsg:
		if msg.Err != nil {
			m.statusMsg = i18n.T(lang, "mpv_stream_err")
			if m.currentScreen == ScreenServers {
				m.servers.playbackErr = fmt.Errorf("%s", i18n.T(lang, "mpv_stream_err"))
			}
		} else {
			m.statusMsg = i18n.T(lang, "playback_finished")
		}
		return m, nil

	// === ROUTING CÁC ASYNC DATA MESSAGES TRỰC TIẾP TỚI SUB-MODEL TƯƠNG ỨNG ===
	case homeItemsLoadedMsg:
		var cmd tea.Cmd
		m.home, cmd = m.home.Update(msg)
		return m, cmd

	case genreResultMsg:
		var cmd tea.Cmd
		m.genre, cmd = m.genre.Update(msg)
		return m, cmd

	case detailLoadedMsg:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd

	case searchResultMsg, searchSuggestionsMsg, debounceSearchMsg:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd

	case backupServersLoadedMsg, playerSourceResolvedMsg:
		var cmd tea.Cmd
		m.servers, cmd = m.servers.Update(msg)
		return m, cmd

	// === SPINNER TICKS: CHỈ FORWARD CHO MODEL ĐANG LOADING ===
	case spinner.TickMsg:
		var tickCmds []tea.Cmd
		if m.home.loading {
			var c tea.Cmd
			m.home, c = m.home.Update(msg)
			tickCmds = append(tickCmds, c)
		}
		if m.genre.loading {
			var c tea.Cmd
			m.genre, c = m.genre.Update(msg)
			tickCmds = append(tickCmds, c)
		}
		if m.detail.loading {
			var c tea.Cmd
			m.detail, c = m.detail.Update(msg)
			tickCmds = append(tickCmds, c)
		}
		if m.search.loading {
			var c tea.Cmd
			m.search, c = m.search.Update(msg)
			tickCmds = append(tickCmds, c)
		}
		if m.servers.state == ServerStateLoading || m.servers.state == ServerStateResolving {
			var c tea.Cmd
			m.servers, c = m.servers.Update(msg)
			tickCmds = append(tickCmds, c)
		}
		return m, tea.Batch(tickCmds...)

	case tea.KeyMsg:
		// Xóa thông báo status cũ khi người dùng bấm phím mới
		m.statusMsg = ""

		// Khi ô tìm kiếm trên cùng (Top Bar) đang được focus
		if m.searchFocused {
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.topSearch.Value())
				m.searchFocused = false
				m.topSearch.Blur()
				if val != "" {
					m.navigateTo(ScreenSearch)
					m.search.Reset()
					m.search.input.SetValue(val)
					m.search.state = SearchStateResults
					m.search.keyword = val
					m.search.currentPage = 1
					m.search.loading = true
					return m, tea.Batch(m.search.spinner.Tick, m.search.fetchSearchResults(val, 1))
				}
				return m, nil

			case "esc":
				m.searchFocused = false
				m.topSearch.Blur()
				return m, nil

			default:
				var tiCmd tea.Cmd
				m.topSearch, tiCmd = m.topSearch.Update(msg)
				return m, tiCmd
			}
		}

		// Xử lý các phím toàn cục khi KHÔNG ở chế độ nhập text
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				return m, tea.Quit
			}

		case "/":
			// Kích hoạt ô tìm kiếm trên cùng
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				m.searchFocused = true
				m.topSearch.Focus()
				return m, textinput.Blink
			}

		case "g":
			// Mở màn hình Thể loại
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				m.navigateTo(ScreenGenre)
				return m, nil
			}

		case "H":
			// Mở màn hình Lịch sử xem phim
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				m.history.Reload()
				m.navigateTo(ScreenHistory)
				return m, nil
			}

		case "c":
			// Mở màn hình Cài đặt cấu hình
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				m.navigateTo(ScreenConfig)
				return m, nil
			}

		case "?":
			if !(m.currentScreen == ScreenSearch && m.search.state == SearchStateInput) {
				m.showHelp = !m.showHelp
				m.help.ShowAll = m.showHelp
				return m, nil
			}

		case "esc":
			if m.showHelp {
				m.showHelp = false
				m.help.ShowAll = false
				return m, nil
			}

			if m.currentScreen == ScreenSearch && m.search.state == SearchStateResults {
				break
			}

			if m.currentScreen != ScreenHome {
				m.navigateBack()
				return m, nil
			}
		}
	}

	// Chuyển tiếp sự kiện phím cho màn hình hiện tại
	var cmd tea.Cmd
	switch m.currentScreen {
	case ScreenHome:
		m.home, cmd = m.home.Update(msg)

	case ScreenSearch:
		m.search, cmd = m.search.Update(msg)

	case ScreenDetail:
		m.detail, cmd = m.detail.Update(msg)

	case ScreenServers:
		m.servers, cmd = m.servers.Update(msg)

	case ScreenGenre:
		m.genre, cmd = m.genre.Update(msg)

	case ScreenHistory:
		m.history, cmd = m.history.Update(msg)

	case ScreenConfig:
		m.config, cmd = m.config.Update(msg)
		// Đồng bộ lại cấu hình và theme ngay khi có thay đổi từ màn hình config
		m.syncConfig()
	}

	return m, cmd
}

// View kết xuất giao diện chính của ứng dụng.
func (m AppModel) View() string {
	lang := m.cfg.Language

	if m.width < 80 || m.height < 24 {
		warning := i18n.T(lang, "screen_small_warn", m.width, m.height)
		return WarningStyle.Render(warning)
	}

	var b strings.Builder

	// 1. App Header + Top Search Bar (Nằm trên cùng)
	leftTitle := AppHeaderStyle.Render(i18n.T(lang, "app_title"))
	searchBarView := "🔍 " + m.topSearch.View()
	if m.searchFocused {
		searchBarView = SelectedItemStyle.Render("🔍 " + m.topSearch.View())
	}

	topNavShortcuts := TabInactiveStyle.Render(fmt.Sprintf("%s  %s  %s  %s",
		i18n.T(lang, "top_shortcut_search"),
		i18n.T(lang, "top_shortcut_genre"),
		i18n.T(lang, "top_shortcut_history"),
		i18n.T(lang, "top_shortcut_config"),
	))
	headerRow := fmt.Sprintf("%s   %s   %s", leftTitle, searchBarView, topNavShortcuts)
	b.WriteString(TruncateDisplay(headerRow, m.width))
	b.WriteString("\n")

	// 2. Nội dung của màn hình hiện tại
	var screenContent string
	switch m.currentScreen {
	case ScreenHome:
		screenContent = m.home.View()
	case ScreenSearch:
		screenContent = m.search.View()
	case ScreenDetail:
		screenContent = m.detail.View()
	case ScreenServers:
		screenContent = m.servers.View()
	case ScreenGenre:
		screenContent = m.genre.View()
	case ScreenHistory:
		screenContent = m.history.View()
	case ScreenConfig:
		screenContent = m.config.View()
	}

	// Bọc khung panel vừa khít 100% chiều rộng và chiều cao terminal
	panelInnerHeight := m.height - 5
	if panelInnerHeight < 10 {
		panelInnerHeight = 10
	}
	panel := PanelStyle.Width(m.width - 2).Height(panelInnerHeight).Render(screenContent)
	b.WriteString(panel)
	b.WriteString("\n")

	// 3. Status bar & Help bar
	status := m.statusMsg
	if status == "" {
		switch m.currentScreen {
		case ScreenHome:
			status = i18n.T(lang, "status_home")
		case ScreenSearch:
			status = i18n.T(lang, "status_search")
		case ScreenDetail:
			status = i18n.T(lang, "status_detail")
		case ScreenServers:
			status = i18n.T(lang, "status_servers")
		case ScreenGenre:
			status = i18n.T(lang, "status_genre")
		case ScreenHistory:
			status = i18n.T(lang, "status_history")
		case ScreenConfig:
			status = i18n.T(lang, "status_config")
		}
	}
	b.WriteString(StatusBarStyle.Render(TruncateDisplay(status, m.width)))
	b.WriteString("\n")

	// Thanh trợ giúp Help
	helpView := m.help.View(m.keys)
	helpLines := strings.Split(helpView, "\n")
	if len(helpLines) > 3 && !m.showHelp {
		helpLines = helpLines[:3]
	}
	for i, line := range helpLines {
		b.WriteString(TruncateDisplay(line, m.width))
		if i < len(helpLines)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
