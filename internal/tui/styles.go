package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	// Bảng màu sắc tương thích Terminal/Dark/Light mode
	ColorPrimary   lipgloss.TerminalColor
	ColorSecondary lipgloss.TerminalColor
	ColorAccent    lipgloss.TerminalColor
	ColorText      lipgloss.TerminalColor
	ColorDim       lipgloss.TerminalColor
	ColorBorder    lipgloss.TerminalColor
	ColorActive    lipgloss.TerminalColor
	ColorError     lipgloss.TerminalColor
	ColorSuccess   lipgloss.TerminalColor

	// Kiểu dáng Header & Tab
	AppHeaderStyle   lipgloss.Style
	TabActiveStyle   lipgloss.Style
	TabInactiveStyle lipgloss.Style

	// Panel & Box
	PanelStyle       lipgloss.Style
	ActivePanelStyle lipgloss.Style

	// Danh sách & Item
	ItemTitleStyle    lipgloss.Style
	SelectedItemStyle lipgloss.Style
	BadgeStyle        lipgloss.Style
	ScoreBadgeStyle   lipgloss.Style

	// Trạng thái & Trợ giúp
	StatusBarStyle lipgloss.Style
	ErrorStyle     lipgloss.Style
	WarningStyle   lipgloss.Style
)

// ApplyTheme cập nhật bảng màu và styles theo chế độ đã cấu hình:
// - "terminal": Sử dụng bảng màu chuẩn ANSI 16 màu của Terminal (Kitty/Foot/Alacritty palette)
// - "dark": Sử dụng bảng màu Dark Mode TrueColor
// - "light": Sử dụng bảng màu Light Mode TrueColor
func ApplyTheme(mode string) {
	switch mode {
	case "dark":
		ColorPrimary = lipgloss.Color("#A388F7")
		ColorSecondary = lipgloss.Color("#5B9CF6")
		ColorAccent = lipgloss.Color("#FFA24B")
		ColorText = lipgloss.Color("#ECEFF4")
		ColorDim = lipgloss.Color("#7E8E9F")
		ColorBorder = lipgloss.Color("#3E4451")
		ColorActive = lipgloss.Color("#43A047")
		ColorError = lipgloss.Color("#F85149")
		ColorSuccess = lipgloss.Color("#56D364")

	case "light":
		ColorPrimary = lipgloss.Color("#7D56F4")
		ColorSecondary = lipgloss.Color("#2E75D3")
		ColorAccent = lipgloss.Color("#E86D22")
		ColorText = lipgloss.Color("#1F2328")
		ColorDim = lipgloss.Color("#656D76")
		ColorBorder = lipgloss.Color("#D0D7DE")
		ColorActive = lipgloss.Color("#0969DA")
		ColorError = lipgloss.Color("#CF222E")
		ColorSuccess = lipgloss.Color("#1A7F37")

	default: // "terminal" (mặc định theo Kitty / Foot)
		// Dùng bảng màu chuẩn ANSI 16 màu (0-15) để tự động hòa hợp với theme terminal
		ColorPrimary = lipgloss.Color("4")   // ANSI Blue
		ColorSecondary = lipgloss.Color("6") // ANSI Cyan
		ColorAccent = lipgloss.Color("3")    // ANSI Yellow / Accent
		ColorText = lipgloss.Color("7")      // ANSI Foreground
		ColorDim = lipgloss.Color("8")       // ANSI Bright Black / Dim
		ColorBorder = lipgloss.Color("8")    // ANSI Bright Black / Border
		ColorActive = lipgloss.Color("2")    // ANSI Green
		ColorError = lipgloss.Color("1")     // ANSI Red
		ColorSuccess = lipgloss.Color("2")   // ANSI Green
	}

	AppHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(ColorPrimary).
		Padding(0, 1)

	TabInactiveStyle = lipgloss.NewStyle().
		Foreground(ColorDim).
		Padding(0, 1)

	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	ActivePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary)

	ItemTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText)

	SelectedItemStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	BadgeStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Padding(0, 1)

	ScoreBadgeStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorError).
		Padding(0, 1)

	WarningStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent).
		Padding(1, 2)
}

func init() {
	ApplyTheme("terminal")
}

func ColorHighlight() lipgloss.TerminalColor {
	return ColorAccent
}

// TruncateDisplay cắt chuỗi an toàn dựa trên độ rộng hiển thị (display width),
// xử lý chuẩn xác ký tự tiếng Việt đa byte, emoji và giữ nguyên vẹn mã màu ANSI escape sequences.
func TruncateDisplay(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	return ansi.Truncate(s, maxWidth, "…")
}

// PadRightDisplay đệm khoảng trắng về bên phải theo đúng độ rộng hiển thị.
func PadRightDisplay(s string, totalWidth int) string {
	w := lipgloss.Width(s)
	if w >= totalWidth {
		return s
	}
	return s + strings.Repeat(" ", totalWidth-w)
}
