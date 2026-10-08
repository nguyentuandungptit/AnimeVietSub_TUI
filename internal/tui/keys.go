package tui

import (
	"avs-tui/internal/i18n"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap định nghĩa toàn bộ phím tắt trong ứng dụng theo chuẩn UI/UX.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Tab         key.Binding
	ShiftTab    key.Binding
	Enter       key.Binding
	Back        key.Binding
	Search      key.Binding
	Reload      key.Binding
	Help        key.Binding
	Quit        key.Binding
	OpenBrowser key.Binding
	Config      key.Binding
	Genre       key.Binding
	History     key.Binding
}

// NewKeyMap tạo KeyMap theo ngôn ngữ cụ thể ("vi" hoặc "en").
func NewKeyMap(lang string) KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", i18n.T(lang, "key_up")),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", i18n.T(lang, "key_down")),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", i18n.T(lang, "key_left")),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", i18n.T(lang, "key_right")),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", i18n.T(lang, "key_tab")),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", i18n.T(lang, "key_shifttab")),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", i18n.T(lang, "key_enter")),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", i18n.T(lang, "key_back")),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", i18n.T(lang, "key_search")),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", i18n.T(lang, "key_reload")),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", i18n.T(lang, "key_help")),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", i18n.T(lang, "key_quit")),
		),
		OpenBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", i18n.T(lang, "key_browser")),
		),
		Config: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", i18n.T(lang, "key_config")),
		),
		Genre: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", i18n.T(lang, "key_genre")),
		),
		History: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", i18n.T(lang, "key_history")),
		),
	}
}

// DefaultKeyMap khởi tạo keymap mặc định.
var DefaultKeyMap = NewKeyMap("vi")

// ShortHelp trả về các phím tắt quan trọng cho thanh trạng thái ngắn.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Enter, k.Back, k.Search, k.Genre, k.History, k.Config, k.Quit}
}

// FullHelp trả về danh sách đầy đủ cho menu trợ giúp.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Tab, k.ShiftTab},
		{k.Enter, k.Back, k.Search, k.Genre, k.History, k.Config},
		{k.Reload, k.OpenBrowser, k.Help, k.Quit},
	}
}
