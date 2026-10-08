package api

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	reFilmID  = regexp.MustCompile(`filmInfo\.filmID\s*=\s*(?:parseInt\(['"]?)?(\d+)`)
	reTitle   = regexp.MustCompile(`filmInfo\.title\s*=\s*["']([^"']+)["']`)
	reFullURL = regexp.MustCompile(`filmInfo\.fullUrl\s*=\s*["']([^"']+)["']`)
)

// ParseEpisodes phân tích danh sách tập phim từ cấu trúc thẻ <ul class="list-episode">.
func ParseEpisodes(htmlContent string) ([]Episode, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var episodes []Episode
	doc.Find("li.episode a.btn-episode, a.episode-link").Each(func(i int, s *goquery.Selection) {
		id := s.AttrOr("data-id", "")
		hash := s.AttrOr("data-hash", "")
		playTech := s.AttrOr("data-play", "")
		href := s.AttrOr("href", "")
		title := s.AttrOr("title", "")
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}

		if id != "" || hash != "" || title != "" {
			episodes = append(episodes, Episode{
				ID:       id,
				Hash:     hash,
				PlayTech: playTech,
				Title:    title,
				URL:      href,
			})
		}
	})

	return episodes, nil
}

// ParseFilmDetail phân tích thông tin chi tiết của phim từ trang /phim/{slug}/.
// Trả về thông tin phim và link tập đầu tiên (firstWatchURL) nếu có.
func ParseFilmDetail(htmlContent string) (*AnimeDetail, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, "", err
	}

	detail := &AnimeDetail{}

	// Trích xuất biến JavaScript toàn cục filmInfo
	if m := reFilmID.FindStringSubmatch(htmlContent); len(m) > 1 {
		detail.FilmID = m[1]
	}
	if m := reTitle.FindStringSubmatch(htmlContent); len(m) > 1 {
		detail.Title = m[1]
	}
	if m := reFullURL.FindStringSubmatch(htmlContent); len(m) > 1 {
		detail.URL = m[1]
	}

	// Fallback tiêu đề nếu chưa có trong filmInfo
	if detail.Title == "" {
		detail.Title = strings.TrimSpace(doc.Find("h1.Title").First().Text())
		if detail.Title == "" {
			detail.Title = strings.TrimSpace(doc.Find("meta[property='og:title']").AttrOr("content", ""))
		}
	}

	// Poster
	detail.Poster = doc.Find("div.Image figure.Objf img").First().AttrOr("src", "")
	if detail.Poster == "" {
		detail.Poster = doc.Find("meta[property='og:image']").AttrOr("content", "")
	}

	// Mô tả
	descEl := doc.Find("div.Description").First()
	if descEl.Length() > 0 {
		detail.Description = strings.TrimSpace(descEl.Text())
	}

	// Điểm đánh giá
	ratingEl := doc.Find(".anime-avg-user-rating").First()
	if ratingEl.Length() > 0 {
		detail.Score = strings.TrimSpace(ratingEl.Text())
	} else {
		detail.Score = strings.TrimSpace(doc.Find("span.Vote").First().Text())
	}

	// Danh sách tập nếu có sẵn trên trang chi tiết
	episodes, _ := ParseEpisodes(htmlContent)
	detail.Episodes = episodes

	// Tìm link tập đầu tiên để chuyển đến trang xem phim nếu chưa có danh sách tập
	var firstWatchURL string
	doc.Find("a[href*='/tap-']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if firstWatchURL == "" && strings.Contains(href, "/tap-") {
			firstWatchURL = href
		}
	})

	return detail, firstWatchURL, nil
}
