package main

import (
	"fmt"
	"os"

	"avs-tui/internal/api"
	"avs-tui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client, err := api.NewClient(api.DefaultBaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi khởi tạo API Client: %v\n", err)
		os.Exit(1)
	}

	app := tui.NewAppModel(client)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Lỗi chạy ứng dụng: %v\n", err)
		os.Exit(1)
	}
}
