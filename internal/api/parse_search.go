package api

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	reBgURL    = regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`)
	rePageOf   = regexp.MustCompile(`Trang\s+(\d+)\s+của\s+(\d+)`)
	rePageLink = regexp.MustCompile(`trang-(\d+)\.html`)
)

// ParseSearchSuggestions phân tích HTML gợi ý tìm kiếm từ POST /ajax/suggest.
func ParseSearchSuggestions(htmlContent string) ([]SearchSuggestion, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var suggestions []SearchSuggestion
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		titleEl := s.Find("a.ss-title").First()
		if titleEl.Length() == 0 {
			return
		}

		title := strings.TrimSpace(titleEl.Text())
		url := titleEl.AttrOr("href", "")

		// Poster từ style background-image
		var poster string
		thumbEl := s.Find("a.thumb").First()
		if style, exists := thumbEl.Attr("style"); exists {
			matches := reBgURL.FindStringSubmatch(style)
			if len(matches) > 1 {
				poster = matches[1]
			}
		}

		// SubText e.g. "Full VietSub"
		subEl := s.Find("div.ss-info p").First()
		subText := strings.TrimSpace(subEl.Text())

		suggestions = append(suggestions, SearchSuggestion{
			Title:   title,
			URL:     url,
			Poster:  poster,
			SubText: subText,
		})
	})

	return suggestions, nil
}

// ParseSearchResults phân tích trang HTML kết quả tìm kiếm đầy đủ từ /tim-kiem/{kw}/...
func ParseSearchResults(htmlContent string) (*SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	result := &SearchResult{
		CurrentPage: 1,
		TotalPages:  1,
	}

	// Danh sách các anime items
	doc.Find("li.TPostMv").Each(func(i int, s *goquery.Selection) {
		item := parseSingleTPostMv(s)
		if item.Title != "" || item.URL != "" {
			result.Items = append(result.Items, item)
		}
	})

	// Phân tích phân trang nếu có
	// Tìm chuỗi "Trang X của Y"
	doc.Find("span").Each(func(i int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		if matches := rePageOf.FindStringSubmatch(txt); len(matches) == 3 {
			if cur, err := strconv.Atoi(matches[1]); err == nil {
				result.CurrentPage = cur
			}
			if total, err := strconv.Atoi(matches[2]); err == nil {
				result.TotalPages = total
			}
		}
	})

	// Kiểm tra trang kế tiếp (Next page)
	if result.CurrentPage < result.TotalPages {
		result.HasNext = true
		// Tìm link tới trang CurrentPage + 1
		nextTarget := "trang-" + strconv.Itoa(result.CurrentPage+1) + ".html"
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href := s.AttrOr("href", "")
			if strings.Contains(href, nextTarget) && result.NextURL == "" {
				result.NextURL = href
			}
		})
	}

	return result, nil
}
