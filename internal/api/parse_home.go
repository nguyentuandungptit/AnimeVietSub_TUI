package api

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseHomeItems phân tích HTML danh sách anime (thẻ li.TPostMv) từ phản hồi /ajax/item.
func ParseHomeItems(htmlContent string) ([]AnimeItem, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var items []AnimeItem
	doc.Find("li.TPostMv").Each(func(i int, s *goquery.Selection) {
		item := parseSingleTPostMv(s)
		if item.Title != "" || item.URL != "" {
			items = append(items, item)
		}
	})

	return items, nil
}

// parseSingleTPostMv trích xuất thông tin từ một thẻ li.TPostMv.
func parseSingleTPostMv(s *goquery.Selection) AnimeItem {
	var item AnimeItem

	article := s.Find("article.TPost")
	if article.Length() > 0 {
		item.ID = article.AttrOr("id", "")
	}

	mainLink := s.Find("article > a").First()
	if mainLink.Length() == 0 {
		mainLink = s.Find("a").First()
	}

	item.URL = mainLink.AttrOr("href", "")

	// Title
	titleEl := s.Find("h2.Title").First()
	if titleEl.Length() == 0 {
		titleEl = s.Find(".Title").First()
	}
	item.Title = strings.TrimSpace(titleEl.Text())

	// Poster
	imgEl := s.Find("div.Image figure.Objf img").First()
	if imgEl.Length() == 0 {
		imgEl = s.Find("img").First()
	}
	item.Poster = imgEl.AttrOr("src", "")

	// Episode status
	epsEl := s.Find("span.mli-eps").First()
	if epsEl.Length() > 0 {
		// e.g. TẬP<i>01</i>
		item.EpisodeStatus = strings.TrimSpace(epsEl.Text())
	}

	// Score
	ratingEl := s.Find(".anime-avg-user-rating").First()
	if ratingEl.Length() > 0 {
		item.Score = strings.TrimSpace(ratingEl.Text())
	} else {
		voteEl := s.Find("span.Vote").First()
		if voteEl.Length() > 0 {
			item.Score = strings.TrimSpace(voteEl.Text())
		}
	}

	// Year / Views
	yearEl := s.Find("span.Year").First()
	if yearEl.Length() > 0 {
		item.Year = strings.TrimSpace(yearEl.Text())
	}

	// Description
	descEl := s.Find("div.Description p").First()
	if descEl.Length() > 0 {
		item.Description = strings.TrimSpace(descEl.Text())
	}

	return item
}
