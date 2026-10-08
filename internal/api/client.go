package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://animevietsub.nl"
	DefaultUA      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	MinReqInterval = 300 * time.Millisecond
	DefaultTimeout = 10 * time.Second
	MaxRetries     = 2
)

// Client cung cấp giao thức giao tiếp HTTP chuẩn với animevietsub.nl.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	mu          sync.Mutex
	lastReqTime time.Time
}

// NewClient tạo Client mới với Cookie Jar và cấu hình mặc định.
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo cookie jar: %w", err)
	}

	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: DefaultTimeout,
		},
	}

	return c, nil
}

// SetHTTPClient cho phép ghi đè http.Client (rất hữu ích cho testing).
func (c *Client) SetHTTPClient(client *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient = client
}

// throttle đảm bảo khoảng cách giữa 2 request liên tiếp luôn >= 300ms.
func (c *Client) throttle(ctx context.Context) error {
	c.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(c.lastReqTime)
	wait := MinReqInterval - elapsed
	if wait > 0 {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		c.mu.Lock()
		c.lastReqTime = time.Now()
		c.mu.Unlock()
		return nil
	}
	c.lastReqTime = now
	c.mu.Unlock()
	return nil
}

// doRequest thực hiện gửi HTTP request với retry (tối đa 2 lần khi lỗi mạng) và rate-limiting.
func (c *Client) doRequest(ctx context.Context, req *ClientRequest) ([]byte, error) {
	// Request object chuẩn bị
	var bodyBytes []byte
	var lastErr error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			// Backoff nhẹ khi retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt*150) * time.Millisecond):
			}
		}

		if err := c.throttle(ctx); err != nil {
			return nil, fmt.Errorf("hủy request trong lúc chờ rate limit: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
		if err != nil {
			return nil, fmt.Errorf("không thể tạo request: %w", err)
		}

		// Gán headers theo mục 4 của docs/API.md
		httpReq.Header.Set("User-Agent", DefaultUA)
		httpReq.Header.Set("Referer", c.baseURL+"/")
		if req.IsAjax {
			httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
			httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue // Thử lại khi lỗi network transport
		}

		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if readErr != nil {
			lastErr = readErr
			continue
		}

		// Với các request nghiệp vụ thông thường, nếu HTTP status code lỗi (>= 400) thì coi là thất bại
		if resp.StatusCode >= 400 && req.URL != c.baseURL+"/" {
			lastErr = fmt.Errorf("HTTP %d (%s)", resp.StatusCode, resp.Status)
			continue
		}

		// Cloudflare warmup: nếu trang đầu trả về 403 để gán cookie,
		// cookiejar đã lưu session cookie. Trong tình huống này nếu chưa qua warmup thì coi như đã nhận cookie.
		bodyBytes = data
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("thực hiện request %s thất bại sau %d lần thử: %w", req.URL, MaxRetries, lastErr)
	}

	return bodyBytes, nil
}

type ClientRequest struct {
	Method string
	URL    string
	Body   string
	IsAjax bool
}

// Warmup thực hiện một request nhẹ vào trang chủ để thu thập cookie phiên ban đầu.
func (c *Client) Warmup(ctx context.Context) error {
	_, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodGet,
		URL:    c.baseURL + "/",
		IsAjax: false,
	})
	return err
}

// GetHomeItems lấy danh sách phim trang chủ theo tab type qua POST /ajax/item.
func (c *Client) GetHomeItems(ctx context.Context, tabType string) ([]AnimeItem, error) {
	form := url.Values{}
	form.Set("widget", "list-film")
	form.Set("type", tabType)

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodPost,
		URL:    c.baseURL + "/ajax/item",
		Body:   form.Encode(),
		IsAjax: true,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách phim trang chủ (tab %s) thất bại: %w", tabType, err)
	}

	return ParseHomeItems(string(data))
}

// GetSearchSuggestions lấy danh sách gợi ý tìm kiếm qua POST /ajax/suggest.
func (c *Client) GetSearchSuggestions(ctx context.Context, keyword string) ([]SearchSuggestion, error) {
	form := url.Values{}
	form.Set("ajaxSearch", "1")
	form.Set("keysearch", keyword)

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodPost,
		URL:    c.baseURL + "/ajax/suggest",
		Body:   form.Encode(),
		IsAjax: true,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy gợi ý tìm kiếm cho '%s' thất bại: %w", keyword, err)
	}

	return ParseSearchSuggestions(string(data))
}

// SearchAnime tìm kiếm phim đầy đủ theo từ khóa và trang qua GET /tim-kiem/{kw}/...
func (c *Client) SearchAnime(ctx context.Context, keyword string, page int) (*SearchResult, error) {
	escaped := url.PathEscape(keyword)
	reqURL := fmt.Sprintf("%s/tim-kiem/%s/", c.baseURL, escaped)
	if page > 1 {
		reqURL = fmt.Sprintf("%s/tim-kiem/%s/trang-%d.html", c.baseURL, escaped, page)
	}

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodGet,
		URL:    reqURL,
		IsAjax: false,
	})
	if err != nil {
		return nil, fmt.Errorf("tìm kiếm phim '%s' (trang %d) thất bại: %w", keyword, page, err)
	}

	return ParseSearchResults(string(data))
}

// GetFilmDetail lấy thông tin chi tiết phim từ link phim.
// Nếu trang chi tiết chưa có danh sách tập, tự động truy cập link xem phim đầu tiên để lấy danh sách tập.
func (c *Client) GetFilmDetail(ctx context.Context, filmURL string) (*AnimeDetail, error) {
	fullURL := filmURL
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = c.baseURL + "/" + strings.TrimLeft(filmURL, "/")
	}

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodGet,
		URL:    fullURL,
		IsAjax: false,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy thông tin chi tiết phim từ %s thất bại: %w", fullURL, err)
	}

	detail, firstWatchURL, err := ParseFilmDetail(string(data))
	if err != nil {
		return nil, fmt.Errorf("phân tích trang chi tiết phim thất bại: %w", err)
	}

	// Nếu trang chi tiết chưa có tập nào và tìm thấy link xem phim, lấy tập từ link đó
	if len(detail.Episodes) == 0 && firstWatchURL != "" {
		watchURL := firstWatchURL
		if !strings.HasPrefix(watchURL, "http") {
			watchURL = c.baseURL + "/" + strings.TrimLeft(watchURL, "/")
		}

		watchData, err := c.doRequest(ctx, &ClientRequest{
			Method: http.MethodGet,
			URL:    watchURL,
			IsAjax: false,
		})
		if err == nil {
			episodes, parseErr := ParseEpisodes(string(watchData))
			if parseErr == nil && len(episodes) > 0 {
				detail.Episodes = episodes
			}
		}
	}

	return detail, nil
}

// GetBackupServers lấy danh sách server dự phòng qua POST /ajax/player với episodeId={id}&backup=1.
func (c *Client) GetBackupServers(ctx context.Context, episodeID string) ([]ServerItem, error) {
	form := url.Values{}
	form.Set("episodeId", episodeID)
	form.Set("backup", "1")

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodPost,
		URL:    c.baseURL + "/ajax/player",
		Body:   form.Encode(),
		IsAjax: true,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy server dự phòng cho tập %s thất bại: %w", episodeID, err)
	}

	return ParseBackupServers(data)
}

// ResolveServerSource lấy nguồn phát khi người dùng chọn server qua POST /ajax/player.
// Body: link={hrefToken}&play={playTech}&id={serverID}&backuplinks=1
func (c *Client) ResolveServerSource(ctx context.Context, hrefToken, playTech, serverID string) (*PlayerResponse, error) {
	form := url.Values{}
	form.Set("link", hrefToken)
	form.Set("play", playTech)
	form.Set("id", serverID)
	form.Set("backuplinks", "1")

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodPost,
		URL:    c.baseURL + "/ajax/player",
		Body:   form.Encode(),
		IsAjax: true,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy nguồn phát từ server ID %s thất bại: %w", serverID, err)
	}

	return ParsePlayerResponse(data)
}

// ResolveEpisodeByHash đổi tập phim qua hash qua POST /ajax/player.
// Body: link={epHash}&id={epID}
func (c *Client) ResolveEpisodeByHash(ctx context.Context, epHash, epID string) (*PlayerResponse, error) {
	form := url.Values{}
	form.Set("link", epHash)
	form.Set("id", epID)

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodPost,
		URL:    c.baseURL + "/ajax/player",
		Body:   form.Encode(),
		IsAjax: true,
	})
	if err != nil {
		return nil, fmt.Errorf("đổi tập qua hash (tập %s) thất bại: %w", epID, err)
	}

	return ParsePlayerResponse(data)
}

// GetGenreItems lấy danh sách phim theo thể loại qua GET /the-loai/{slug}/ hoặc phân trang.
func (c *Client) GetGenreItems(ctx context.Context, genreSlug string, page int) (*SearchResult, error) {
	reqURL := fmt.Sprintf("%s/the-loai/%s/", c.baseURL, genreSlug)
	if page > 1 {
		reqURL = fmt.Sprintf("%s/the-loai/%s/trang-%d.html", c.baseURL, genreSlug, page)
	}

	data, err := c.doRequest(ctx, &ClientRequest{
		Method: http.MethodGet,
		URL:    reqURL,
		IsAjax: false,
	})
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách thể loại '%s' (trang %d) thất bại: %w", genreSlug, page, err)
	}

	return ParseSearchResults(string(data))
}
