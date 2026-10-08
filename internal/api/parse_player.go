package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type backupResponse struct {
	Success int    `json:"success"`
	HTML    string `json:"html"`
}

// ParseBackupServers phân tích phản hồi JSON từ POST /ajax/player?backup=1 để lấy danh sách server.
func ParseBackupServers(data []byte) ([]ServerItem, error) {
	var resp backupResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("không thể giải mã JSON backup servers: %w", err)
	}

	if resp.HTML == "" {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.HTML))
	if err != nil {
		return nil, fmt.Errorf("không thể phân tích HTML backup servers: %w", err)
	}

	var servers []ServerItem
	doc.Find("a.btn3dsv").Each(func(i int, s *goquery.Selection) {
		id := s.AttrOr("data-id", "")
		playTech := s.AttrOr("data-play", "")
		href := s.AttrOr("data-href", "")
		name := strings.TrimSpace(s.Text())
		active := s.HasClass("active")

		servers = append(servers, ServerItem{
			ID:        id,
			Name:      name,
			PlayTech:  playTech,
			HrefToken: href,
			Active:    active,
		})
	})

	return servers, nil
}

// ParsePlayerResponse phân tích phản hồi JSON từ /ajax/player khi lấy nguồn phát video.
func ParsePlayerResponse(data []byte) (*PlayerResponse, error) {
	var resp PlayerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("không thể giải mã phản hồi Player: %w", err)
	}
	return &resp, nil
}
