package tui

import (
	"fmt"
	"strings"

	"avs-tui/internal/config"
	"avs-tui/internal/history"
	"avs-tui/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigModel quản lý màn hình cài đặt cấu hình ứng dụng.
type ConfigModel struct {
	cfg          *config.Config
	cursor       int
	confirmClear bool
	statusMsg    string
	width        int
	height       int
}

// NewConfigModel khởi tạo ConfigModel.
func NewConfigModel() ConfigModel {
	return ConfigModel{
		cfg:          config.LoadConfig(),
		cursor:       0,
		confirmClear: false,
	}
}

// SetSize cập nhật kích thước khung nhìn.
func (m *ConfigModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update xử lý sự kiện trong màn hình cấu hình.
func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.confirmClear = false
			m.statusMsg = ""
			return m, nil

		case "down", "j":
			if m.cursor < 5 {
				m.cursor++
			}
			m.confirmClear = false
			m.statusMsg = ""
			return m, nil

		case "right", "l":
			if m.cursor != 5 {
				m.confirmClear = false
				m.rotateOptionNext()
				_ = config.SaveConfig(m.cfg)
			}
			return m, nil

		case "left", "h":
			if m.cursor != 5 {
				m.confirmClear = false
				m.rotateOptionPrev()
				_ = config.SaveConfig(m.cfg)
			}
			return m, nil

		case "enter":
			if m.cursor == 5 {
				if !m.confirmClear {
					m.confirmClear = true
					m.statusMsg = i18n.T(m.cfg.Language, "cfg_confirm_clear")
				} else {
					_ = history.ClearHistory()
					m.confirmClear = false
					m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_cleared")
				}
			} else {
				m.confirmClear = false
				m.rotateOptionNext()
				_ = config.SaveConfig(m.cfg)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *ConfigModel) rotateOptionNext() {
	switch m.cursor {
	case 0: // Ngôn ngữ
		if m.cfg.Language == "vi" {
			m.cfg.Language = "en"
		} else {
			m.cfg.Language = "vi"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_lang")

	case 1: // Chế độ phát video
		switch m.cfg.PlayerMode {
		case "auto":
			m.cfg.PlayerMode = "app_window"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_app")
		case "app_window":
			m.cfg.PlayerMode = "browser"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_browser")
		case "browser":
			m.cfg.PlayerMode = "mpv"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_mpv")
		default:
			m.cfg.PlayerMode = "auto"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_auto")
		}

	case 2: // Tự động mở trình phát khi chọn tập
		m.cfg.AutoOpenBrowserForIframe = !m.cfg.AutoOpenBrowserForIframe
		if m.cfg.AutoOpenBrowserForIframe {
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_ap_on")
		} else {
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_ap_off")
		}

	case 3: // Chất lượng video ưu tiên
		switch m.cfg.PreferredQuality {
		case "1080p":
			m.cfg.PreferredQuality = "720p"
		case "720p":
			m.cfg.PreferredQuality = "auto"
		default:
			m.cfg.PreferredQuality = "1080p"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_quality", m.cfg.PreferredQuality)

	case 4: // Theme Mode
		switch m.cfg.ThemeMode {
		case "terminal":
			m.cfg.ThemeMode = "dark"
		case "dark":
			m.cfg.ThemeMode = "light"
		default:
			m.cfg.ThemeMode = "terminal"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_theme", m.cfg.ThemeMode)
	}
}

func (m *ConfigModel) rotateOptionPrev() {
	switch m.cursor {
	case 0: // Ngôn ngữ
		if m.cfg.Language == "vi" {
			m.cfg.Language = "en"
		} else {
			m.cfg.Language = "vi"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_lang")

	case 1: // Chế độ phát video (đảo ngược chiều)
		switch m.cfg.PlayerMode {
		case "mpv":
			m.cfg.PlayerMode = "browser"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_browser")
		case "browser":
			m.cfg.PlayerMode = "app_window"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_app")
		case "app_window":
			m.cfg.PlayerMode = "auto"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_auto")
		default:
			m.cfg.PlayerMode = "mpv"
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_pm_mpv")
		}

	case 2: // Tự động mở trình phát khi chọn tập
		m.cfg.AutoOpenBrowserForIframe = !m.cfg.AutoOpenBrowserForIframe
		if m.cfg.AutoOpenBrowserForIframe {
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_ap_on")
		} else {
			m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_ap_off")
		}

	case 3: // Chất lượng video ưu tiên (đảo ngược chiều)
		switch m.cfg.PreferredQuality {
		case "auto":
			m.cfg.PreferredQuality = "720p"
		case "720p":
			m.cfg.PreferredQuality = "1080p"
		default:
			m.cfg.PreferredQuality = "auto"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_quality", m.cfg.PreferredQuality)

	case 4: // Theme Mode (đảo ngược chiều)
		switch m.cfg.ThemeMode {
		case "light":
			m.cfg.ThemeMode = "dark"
		case "dark":
			m.cfg.ThemeMode = "terminal"
		default:
			m.cfg.ThemeMode = "light"
		}
		m.statusMsg = i18n.T(m.cfg.Language, "cfg_msg_theme", m.cfg.ThemeMode)
	}
}

// View kết xuất màn hình cấu hình.
func (m ConfigModel) View() string {
	lang := m.cfg.Language
	var b strings.Builder

	header := AppHeaderStyle.Render(i18n.T(lang, "cfg_header"))
	b.WriteString(header)
	b.WriteString("\n\n")

	contentWidth := m.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	options := []struct {
		label string
		value string
	}{
		{
			label: i18n.T(lang, "cfg_lang"),
			value: i18n.T(lang, "cfg_lang_val"),
		},
		{
			label: i18n.T(lang, "cfg_player_mode"),
			value: map[string]string{
				"auto":       i18n.T(lang, "cfg_pm_auto"),
				"app_window": i18n.T(lang, "cfg_pm_app"),
				"browser":    i18n.T(lang, "cfg_pm_browser"),
				"mpv":        i18n.T(lang, "cfg_pm_mpv"),
			}[m.cfg.PlayerMode],
		},
		{
			label: i18n.T(lang, "cfg_auto_play"),
			value: map[bool]string{
				true:  i18n.T(lang, "cfg_auto_play_on"),
				false: i18n.T(lang, "cfg_auto_play_off"),
			}[m.cfg.AutoOpenBrowserForIframe],
		},
		{
			label: i18n.T(lang, "cfg_quality"),
			value: strings.ToUpper(m.cfg.PreferredQuality),
		},
		{
			label: i18n.T(lang, "cfg_theme"),
			value: map[string]string{
				"terminal": i18n.T(lang, "cfg_theme_term"),
				"dark":     i18n.T(lang, "cfg_theme_dark"),
				"light":    i18n.T(lang, "cfg_theme_light"),
			}[m.cfg.ThemeMode],
		},
		{
			label: i18n.T(lang, "cfg_clear_history"),
			value: func() string {
				if m.confirmClear {
					return i18n.T(lang, "cfg_confirm_clear")
				}
				return i18n.T(lang, "cfg_clear_btn")
			}(),
		},
	}

	for i, opt := range options {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}

		line := fmt.Sprintf("%s%-36s : %s", prefix, opt.label, opt.value)
		if i == m.cursor {
			b.WriteString(SelectedItemStyle.Render(TruncateDisplay(line, contentWidth)))
		} else {
			b.WriteString(TruncateDisplay(line, contentWidth))
		}
		b.WriteString("\n\n")
	}

	if m.statusMsg != "" {
		b.WriteString(ScoreBadgeStyle.Render(m.statusMsg))
		b.WriteString("\n\n")
	}

	b.WriteString(TabInactiveStyle.Render(i18n.T(lang, "cfg_footer")))

	return b.String()
}
