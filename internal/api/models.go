package api

import (
	"encoding/json"
	"fmt"
)

// HomeTab đại diện cho một tab lọc phim trên trang chủ.
type HomeTab struct {
	Type  string
	Label string
}

// HomeTabs danh sách 11 tabs trang chủ theo mục 3.2.B của docs/API.md.
var HomeTabs = []HomeTab{
	{Type: "anime-new", Label: "Mới cập nhật"},
	{Type: "anime-season", Label: "Theo mùa"},
	{Type: "anime-series", Label: "Anime bộ"},
	{Type: "anime-single", Label: "Anime lẻ"},
	{Type: "hh-trungquoc", Label: "HH Trung Quốc"},
	{Type: "hot-viewed-today", Label: "Xem nhiều hôm nay"},
	{Type: "hot-viewed-season", Label: "Xem nhiều mùa này"},
	{Type: "hot-top-voted", Label: "Yêu thích nhất"},
	{Type: "hot-viewed-month", Label: "Xem nhiều tháng"},
	{Type: "top-bo-week", Label: "Top bộ tuần"},
	{Type: "top-le-week", Label: "Top lẻ tuần"},
}

// GenreItem thông tin một thể loại anime.
type GenreItem struct {
	Slug string
	Name string
}

// DefaultGenres danh sách các thể loại anime trên animevietsub.nl.
var DefaultGenres = []GenreItem{
	{Slug: "hanh-dong", Name: "Hành Động"},
	{Slug: "phieu-luu", Name: "Phiêu Lưu"},
	{Slug: "hai-huoc", Name: "Hài Hước"},
	{Slug: "phep-thuat", Name: "Phép Thuật"},
	{Slug: "truong-hoc", Name: "Trường Học"},
	{Slug: "doi-thuong", Name: "Đời Thường"},
	{Slug: "drama", Name: "Drama"},
	{Slug: "kinh-di", Name: "Kinh Dị"},
	{Slug: "tinh-cam", Name: "Tình Cảm"},
	{Slug: "the-thao", Name: "Thể Thao"},
	{Slug: "co-trang", Name: "Cổ Trang"},
	{Slug: "vien-tuong", Name: "Viễn Tưởng"},
	{Slug: "bi-an", Name: "Bí Ẩn"},
	{Slug: "sieu-nhien", Name: "Siêu Nhiên"},
}

// AnimeItem đại diện cho một anime trong danh sách (trang chủ hoặc tìm kiếm).
type AnimeItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	Poster        string `json:"poster"`
	EpisodeStatus string `json:"episode_status"`
	Score         string `json:"score"`
	Year          string `json:"year"`
	Description   string `json:"description"`
}

// SearchSuggestion mục gợi ý tìm kiếm nhanh từ /ajax/suggest.
type SearchSuggestion struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Poster  string `json:"poster"`
	SubText string `json:"sub_text"`
}

// SearchResult kết quả tìm kiếm đầy đủ từ /tim-kiem/{kw}/ kèm phân trang.
type SearchResult struct {
	Items       []AnimeItem
	CurrentPage int
	TotalPages  int
	HasNext     bool
	NextURL     string
}

// AnimeDetail thông tin chi tiết một bộ phim.
type AnimeDetail struct {
	FilmID      string    `json:"film_id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Poster      string    `json:"poster"`
	Description string    `json:"description"`
	Score       string    `json:"score"`
	Genres      []string  `json:"genres"`
	Episodes    []Episode `json:"episodes"`
}

// Episode thông tin một tập phim.
type Episode struct {
	ID       string `json:"id"`
	Hash     string `json:"hash"`
	PlayTech string `json:"play_tech"`
	Title    string `json:"title"`
	URL      string `json:"url"`
}

// ServerItem thông tin một server phát video.
type ServerItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PlayTech  string `json:"play_tech"`
	HrefToken string `json:"href_token"`
	Active    bool   `json:"active"`
}

// StreamSource nguồn phát stream (khi playTech == "api").
type StreamSource struct {
	File  string `json:"file"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// PlayerResponse phản hồi từ /ajax/player khi resolve link phát.
type PlayerResponse struct {
	FxStatus   int            `json:"_fxStatus"`
	Success    int            `json:"success"`
	Title      string         `json:"title"`
	PlayTech   string         `json:"playTech"`
	SingleLink string         `json:"-"`
	MultiLinks []StreamSource `json:"-"`
	RawLink    json.RawMessage `json:"link"`
}

// UnmarshalJSON giải mã tuỳ biến trường "link" có thể là chuỗi hoặc mảng object.
func (p *PlayerResponse) UnmarshalJSON(data []byte) error {
	type Alias PlayerResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(p.RawLink) == 0 {
		return nil
	}

	// Thử parse thành chuỗi string (dành cho iframe / embed)
	var single string
	if err := json.Unmarshal(p.RawLink, &single); err == nil {
		p.SingleLink = single
		return nil
	}

	// Thử parse thành mảng StreamSource (dành cho api)
	var multi []StreamSource
	if err := json.Unmarshal(p.RawLink, &multi); err == nil {
		p.MultiLinks = multi
		return nil
	}

	return fmt.Errorf("không thể phân tích trường 'link' trong PlayerResponse: %s", string(p.RawLink))
}
