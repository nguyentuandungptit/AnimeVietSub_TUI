package player

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// FinishedMsg gửi về Bubble Tea update loop khi mpv thoát.
type FinishedMsg struct {
	Err error
}

// ValidateStreamURL kiểm tra URL có scheme hợp lệ (http/https) và có host không rỗng.
func ValidateStreamURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL không hợp lệ: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("chỉ chấp nhận giao thức http hoặc https, nhận được: %s", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL không hợp lệ (thiếu host)")
	}
	return nil
}

// IsMPVInstalled kiểm tra xem binary mpv đã cài trên hệ thống chưa.
func IsMPVInstalled() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

// GetInstallInstructions trả về hướng dẫn cài đặt mpv.
func GetInstallInstructions() string {
	switch runtime.GOOS {
	case "linux":
		return "mpv chưa được cài đặt. Bạn có thể cài đặt bằng lệnh:\n  - Ubuntu/Debian: sudo apt install mpv\n  - Fedora: sudo dnf install mpv\n  - Arch Linux: sudo pacman -S mpv"
	case "darwin":
		return "mpv chưa được cài đặt. Bạn có thể cài đặt bằng lệnh:\n  brew install mpv"
	case "windows":
		return "mpv chưa được cài đặt. Vui lòng tải mpv từ https://mpv.io hoặc cài qua winget/scoop:\n  winget install shinchiro.mpv"
	default:
		return "mpv chưa được cài đặt. Vui lòng cài đặt mpv từ https://mpv.io"
	}
}

// BuildMPVCommand xây dựng lệnh mpv với các tham số tối ưu và header cần thiết.
// Đã loại bỏ --http-header-fields để tránh lỗi ngắt header khi User-Agent chứa dấu phẩy.
// Đặt flag "--" trước URL để ngăn injection tham số.
func BuildMPVCommand(videoURL, title string) *exec.Cmd {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	referrer := "https://animevietsub.nl/"

	args := []string{
		"--really-quiet",
		fmt.Sprintf("--user-agent=%s", ua),
		fmt.Sprintf("--referrer=%s", referrer),
	}

	if title != "" {
		args = append(args, fmt.Sprintf("--title=AnimeVietSub - %s", title))
	}

	args = append(args, "--", videoURL)
	return exec.Command("mpv", args...)
}

// PlayStream trả về tea.Cmd thực thi mpv qua tea.ExecProcess sau khi kiểm tra URL.
func PlayStream(videoURL, title string) tea.Cmd {
	if err := ValidateStreamURL(videoURL); err != nil {
		return func() tea.Msg {
			return FinishedMsg{Err: err}
		}
	}
	c := BuildMPVCommand(videoURL, title)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return FinishedMsg{Err: err}
	})
}

// OpenBrowser mở URL bằng trình duyệt mặc định của hệ điều hành.
func OpenBrowser(targetURL string) error {
	if err := ValidateStreamURL(targetURL); err != nil {
		return err
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		return fmt.Errorf("hệ điều hành không được hỗ trợ để mở trình duyệt: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

// FindChromiumBinary tìm binary của trình duyệt nhân Chromium hỗ trợ chế độ --app.
func FindChromiumBinary() string {
	candidates := []string{
		"chromium",
		"google-chrome",
		"google-chrome-stable",
		"brave-browser",
		"brave",
		"microsoft-edge",
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// PlayStandaloneWebPlayer mở URL dưới dạng cửa sổ ứng dụng Web Player độc lập (không address bar/tabs).
// Nếu không tìm thấy Chromium-based browser, tự động fallback sang OpenBrowser.
func PlayStandaloneWebPlayer(targetURL string) (string, error) {
	if err := ValidateStreamURL(targetURL); err != nil {
		return "", err
	}

	chromePath := FindChromiumBinary()
	if chromePath != "" {
		cmd := exec.Command(chromePath, fmt.Sprintf("--app=%s", targetURL), "--window-size=1280,720")
		if err := cmd.Start(); err == nil {
			go func() {
				_ = cmd.Wait()
			}()
			return "Cửa sổ Standalone Web Player", nil
		}
	}

	// Fallback sang trình duyệt mặc định
	if err := OpenBrowser(targetURL); err != nil {
		return "", err
	}
	return "Trình duyệt mặc định", nil
}

