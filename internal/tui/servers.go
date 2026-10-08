package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"avs-tui/internal/api"
	"avs-tui/internal/config"
	"avs-tui/internal/history"
	"avs-tui/internal/i18n"
	"avs-tui/internal/player"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PlayVideoMsg yêu cầu root model app.go khởi chạy trình phát video qua tea.ExecProcess.
type PlayVideoMsg struct {
	VideoURL string
	Title    string
	PlayTech string
}

type backupServersLoadedMsg struct {
	servers []api.ServerItem
	err     error
}

type playerSourceResolvedMsg struct {
	response *api.PlayerResponse
	err      error
}

// ServerState trạng thái của quy trình chọn server và chất lượng.
type ServerState int

const (
	ServerStateLoading ServerState = iota
	ServerStateSelectServer
	ServerStateResolving
	ServerStateSelectQuality
	ServerStatePlaybackInfo
)

// ServersModel quản lý chọn server phát, chọn độ phân giải và thông tin stream.
type ServersModel struct {
	client       *api.Client
	state        ServerState
	episode      api.Episode
	animeTitle   string
	animeURL     string
	servers      []api.ServerItem
	serverCursor int
	resolvedResp *api.PlayerResponse
	qualityList  []api.StreamSource
	qualCursor   int
	finalURL     string
	playbackErr  error
	launchNotice string
	spinner      spinner.Model
	width        int
	height       int
}

// NewServersModel khởi tạo ServersModel.
func NewServersModel(client *api.Client) ServersModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return ServersModel{
		client:  client,
		state:   ServerStateLoading,
		spinner: s,
	}
}

// Init khởi động spinner.
func (m ServersModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// SetSize cập nhật kích thước khung nhìn.
func (m *ServersModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// LoadServers bắt đầu quy trình lấy server dự phòng cho một tập phim.
func (m *ServersModel) LoadServers(animeTitle string, animeURL string, ep api.Episode) tea.Cmd {
	m.animeTitle = animeTitle
	m.animeURL = animeURL
	m.episode = ep
	m.state = ServerStateLoading
	m.servers = nil
	m.serverCursor = 0
	m.resolvedResp = nil
	m.qualityList = nil
	m.qualCursor = 0
	m.finalURL = ""
	m.playbackErr = nil
	m.launchNotice = ""

	return tea.Batch(
		m.spinner.Tick,
		m.fetchBackupServers(ep.ID),
	)
}

func (m ServersModel) fetchBackupServers(epID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		servers, err := m.client.GetBackupServers(ctx, epID)
		return backupServersLoadedMsg{
			servers: servers,
			err:     err,
		}
	}
}

func (m ServersModel) resolveServer(server api.ServerItem) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := m.client.ResolveServerSource(ctx, server.HrefToken, server.PlayTech, server.ID)
		return playerSourceResolvedMsg{
			response: resp,
			err:      err,
		}
	}
}

func (m *ServersModel) recordHistory() {
	animeURL := m.animeURL
	if animeURL == "" {
		animeURL = m.episode.URL
	}
	_ = history.AddHistory(history.HistoryItem{
		AnimeTitle:   m.animeTitle,
		AnimeURL:     animeURL,
		EpisodeTitle: m.episode.Title,
		EpisodeID:    m.episode.ID,
		EpisodeHash:  m.episode.Hash,
	})
}

// Update xử lý sự kiện trong quy trình chọn server.
func (m ServersModel) Update(msg tea.Msg) (ServersModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.state == ServerStateLoading || m.state == ServerStateResolving {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case backupServersLoadedMsg:
		if msg.err != nil {
			m.state = ServerStatePlaybackInfo
			m.playbackErr = fmt.Errorf("không thể lấy danh sách server: %w", msg.err)
			return m, nil
		}

		if len(msg.servers) == 0 {
			// Nếu không có backup server, thử dùng hash của tập
			if m.episode.Hash != "" {
				m.state = ServerStateResolving
				return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					resp, err := m.client.ResolveEpisodeByHash(ctx, m.episode.Hash, m.episode.ID)
					return playerSourceResolvedMsg{response: resp, err: err}
				})
			}
			m.state = ServerStatePlaybackInfo
			m.playbackErr = fmt.Errorf("không tìm thấy nguồn server phát cho tập này")
			return m, nil
		}

		m.servers = msg.servers
		m.serverCursor = 0
		m.state = ServerStateSelectServer
		return m, nil

	case playerSourceResolvedMsg:
		if msg.err != nil {
			m.state = ServerStatePlaybackInfo
			m.playbackErr = fmt.Errorf("lỗi resolve nguồn phát: %w", msg.err)
			return m, nil
		}

		m.resolvedResp = msg.response
		tech := msg.response.PlayTech
		cfg := config.LoadConfig()
		lang := cfg.Language

		// Trường hợp 1: API -> Chọn chất lượng nếu có nhiều nguồn hoặc tự chọn theo cấu hình
		if tech == "api" && len(msg.response.MultiLinks) > 0 {
			m.qualityList = msg.response.MultiLinks
			m.qualCursor = 0

			if len(m.qualityList) == 1 {
				// Chỉ có 1 chất lượng, phát luôn và đặt trạng thái ServerStatePlaybackInfo
				m.finalURL = m.qualityList[0].File
				m.state = ServerStatePlaybackInfo
				m.recordHistory()
				return m, func() tea.Msg {
					return PlayVideoMsg{
						VideoURL: m.finalURL,
						Title:    fmt.Sprintf("%s - %s", m.animeTitle, m.episode.Title),
						PlayTech: tech,
					}
				}
			}

			// Nếu có cấu hình PreferredQuality khác "auto", kiểm tra xem có source nào khớp không
			if cfg.PreferredQuality != "auto" {
				matchIdx := -1
				targetQual := strings.ToLower(cfg.PreferredQuality)
				for i, src := range m.qualityList {
					if strings.Contains(strings.ToLower(src.Label), targetQual) {
						matchIdx = i
						break
					}
				}
				if matchIdx >= 0 {
					m.qualCursor = matchIdx
					m.finalURL = m.qualityList[matchIdx].File
					m.state = ServerStatePlaybackInfo
					m.recordHistory()
					return m, func() tea.Msg {
						return PlayVideoMsg{
							VideoURL: m.finalURL,
							Title:    fmt.Sprintf("%s - %s (%s)", m.animeTitle, m.episode.Title, m.qualityList[matchIdx].Label),
							PlayTech: tech,
						}
					}
				}
			}

			m.state = ServerStateSelectQuality
			return m, nil
		}

		// Tự động ghi lại lịch sử xem cho Iframe hoặc Embed
		m.recordHistory()

		// Trường hợp 2: Iframe hoặc Embed -> Hiển thị thông tin link để mở trình duyệt hoặc thử MPV
		link := msg.response.SingleLink
		if link == "" && len(msg.response.MultiLinks) > 0 {
			link = msg.response.MultiLinks[0].File
		}

		m.finalURL = link
		m.state = ServerStatePlaybackInfo

		if cfg.AutoOpenBrowserForIframe && link != "" {
			switch cfg.PlayerMode {
			case "mpv":
				return m, func() tea.Msg {
					return PlayVideoMsg{
						VideoURL: link,
						Title:    fmt.Sprintf("%s - %s", m.animeTitle, m.episode.Title),
						PlayTech: tech,
					}
				}
			case "browser":
				if err := player.OpenBrowser(link); err == nil {
					m.launchNotice = i18n.T(lang, "servers_opened_br")
				} else {
					m.playbackErr = fmt.Errorf("không thể mở trình duyệt: %w", err)
				}
			default: // "auto" hoặc "app_window"
				method, err := player.PlayStandaloneWebPlayer(link)
				if err == nil {
					m.launchNotice = i18n.T(lang, "servers_auto_open", method)
				} else {
					m.playbackErr = fmt.Errorf("không thể khởi chạy trình phát: %w", err)
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case ServerStateSelectServer:
			switch msg.String() {
			case "up", "k":
				if m.serverCursor > 0 {
					m.serverCursor--
				}
				return m, nil

			case "down", "j":
				if m.serverCursor < len(m.servers)-1 {
					m.serverCursor++
				}
				return m, nil

			case "enter":
				if len(m.servers) > 0 && m.serverCursor < len(m.servers) {
					selected := m.servers[m.serverCursor]
					m.state = ServerStateResolving
					return m, tea.Batch(m.spinner.Tick, m.resolveServer(selected))
				}
			}

		case ServerStateSelectQuality:
			switch msg.String() {
			case "up", "k":
				if m.qualCursor > 0 {
					m.qualCursor--
				}
				return m, nil

			case "down", "j":
				if m.qualCursor < len(m.qualityList)-1 {
					m.qualCursor++
				}
				return m, nil

			case "enter":
				if len(m.qualityList) > 0 && m.qualCursor < len(m.qualityList) {
					m.finalURL = m.qualityList[m.qualCursor].File
					m.state = ServerStatePlaybackInfo
					m.recordHistory()
					return m, func() tea.Msg {
						return PlayVideoMsg{
							VideoURL: m.finalURL,
							Title:    fmt.Sprintf("%s - %s (%s)", m.animeTitle, m.episode.Title, m.qualityList[m.qualCursor].Label),
							PlayTech: "api",
						}
					}
				}
			}

		case ServerStatePlaybackInfo:
			cfg := config.LoadConfig()
			lang := cfg.Language

			switch msg.String() {
			case "enter":
				if m.finalURL != "" {
					tech := "embed"
					if m.resolvedResp != nil && m.resolvedResp.PlayTech != "" {
						tech = m.resolvedResp.PlayTech
					}
					m.recordHistory()

					switch cfg.PlayerMode {
					case "mpv":
						return m, func() tea.Msg {
							return PlayVideoMsg{
								VideoURL: m.finalURL,
								Title:    fmt.Sprintf("%s - %s", m.animeTitle, m.episode.Title),
								PlayTech: tech,
							}
						}
					case "browser":
						if err := player.OpenBrowser(m.finalURL); err == nil {
							m.playbackErr = nil
							m.launchNotice = i18n.T(lang, "servers_opened_br")
						} else {
							m.playbackErr = fmt.Errorf("không thể mở trình duyệt: %w", err)
						}
						return m, nil
					default: // "auto" hoặc "app_window"
						method, err := player.PlayStandaloneWebPlayer(m.finalURL)
						if err == nil {
							m.playbackErr = nil
							m.launchNotice = i18n.T(lang, "servers_launching", method)
						} else {
							m.playbackErr = fmt.Errorf("lỗi khởi chạy: %w", err)
						}
						return m, nil
					}
				}

			case "m":
				// Phím 'm': Cố gắng phát bằng MPV
				if m.finalURL != "" {
					m.recordHistory()
					return m, func() tea.Msg {
						return PlayVideoMsg{
							VideoURL: m.finalURL,
							Title:    fmt.Sprintf("%s - %s", m.animeTitle, m.episode.Title),
							PlayTech: "embed",
						}
					}
				}

			case "o":
				// Mở liên kết qua trình duyệt web mặc định
				if m.finalURL != "" {
					if err := player.OpenBrowser(m.finalURL); err == nil {
						m.playbackErr = nil
						m.launchNotice = i18n.T(lang, "servers_opened_br")
					} else {
						m.playbackErr = err
					}
				}
				return m, nil

			case "r":
				return m, m.LoadServers(m.animeTitle, m.animeURL, m.episode)
			}
		}
	}

	return m, nil
}

// View kết xuất màn hình server & playback.
func (m ServersModel) View() string {
	var b strings.Builder

	contentWidth := m.width - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	cfg := config.LoadConfig()
	lang := cfg.Language

	header := fmt.Sprintf("📺 %s - %s", m.animeTitle, m.episode.Title)
	b.WriteString(AppHeaderStyle.Render(TruncateDisplay(header, contentWidth)))
	b.WriteString("\n\n")

	switch m.state {
	case ServerStateLoading:
		b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "servers_loading")))

	case ServerStateResolving:
		b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), i18n.T(lang, "servers_resolving")))

	case ServerStateSelectServer:
		b.WriteString(TabActiveStyle.Render(i18n.T(lang, "servers_select")))
		b.WriteString("\n\n")

		for i, s := range m.servers {
			prefix := "  "
			if i == m.serverCursor {
				prefix = "> "
			}
			badge := fmt.Sprintf("[%s]", strings.ToUpper(s.PlayTech))
			name := s.Name
			if s.Active {
				name += " " + i18n.T(lang, "servers_default")
			}
			line := fmt.Sprintf("%s%-10s %s", prefix, badge, name)

			if i == m.serverCursor {
				b.WriteString(SelectedItemStyle.Render(TruncateDisplay(line, contentWidth)))
			} else {
				b.WriteString(TruncateDisplay(line, contentWidth))
			}
			b.WriteString("\n")
		}

	case ServerStateSelectQuality:
		b.WriteString(TabActiveStyle.Render(i18n.T(lang, "servers_quality")))
		b.WriteString("\n\n")

		for i, q := range m.qualityList {
			prefix := "  "
			if i == m.qualCursor {
				prefix = "> "
			}
			line := fmt.Sprintf("%s%s (%s)", prefix, q.Label, q.Type)

			if i == m.qualCursor {
				b.WriteString(SelectedItemStyle.Render(TruncateDisplay(line, contentWidth)))
			} else {
				b.WriteString(TruncateDisplay(line, contentWidth))
			}
			b.WriteString("\n")
		}

	case ServerStatePlaybackInfo:
		if m.launchNotice != "" {
			b.WriteString(ScoreBadgeStyle.Render(m.launchNotice))
			b.WriteString("\n\n")
		}

		if m.playbackErr != nil {
			b.WriteString(ErrorStyle.Render(fmt.Sprintf("Notice: %v", m.playbackErr)))
			b.WriteString("\n\n")
		}

		if !player.IsMPVInstalled() {
			b.WriteString(WarningStyle.Render(player.GetInstallInstructions()))
			b.WriteString("\n\n")
		}

		if m.finalURL != "" {
			techName := "VIDEO"
			if m.resolvedResp != nil && m.resolvedResp.PlayTech != "" {
				techName = strings.ToUpper(m.resolvedResp.PlayTech)
			}
			b.WriteString(TabActiveStyle.Render(i18n.T(lang, "servers_source", techName)))
			b.WriteString("\n")
			b.WriteString(TruncateDisplay("URL: "+m.finalURL, contentWidth))
			b.WriteString("\n\n")

			if techName == "IFRAME" || techName == "EMBED" {
				b.WriteString(SelectedItemStyle.Render(i18n.T(lang, "servers_play_enter")))
				b.WriteString("\n")
				b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "servers_open_brow")))
				b.WriteString("\n")
				b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "servers_try_mpv")))
			} else {
				b.WriteString(SelectedItemStyle.Render(i18n.T(lang, "servers_play_mpv")))
				b.WriteString("\n")
				b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "servers_open_brow")))
			}
			b.WriteString("\n")
			b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "servers_back")))
		}
	}

	return b.String()
}
