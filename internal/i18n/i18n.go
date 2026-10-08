package i18n

import (
	"fmt"
	"time"
)

// Bảng từ điển dịch đa ngôn ngữ cho ứng dụng.
var translations = map[string]map[string]string{
	"vi": {
		// App Chrome & Navigation
		"app_title":            "📺 ANIMEVIETSUB TUI",
		"top_search_hint":      "Nhấn '/' để tìm kiếm nhanh...",
		"top_shortcut_search":  "[/] Tìm",
		"top_shortcut_genre":   "[g] Thể loại",
		"top_shortcut_history": "[H] Lịch sử",
		"top_shortcut_config":  "[c] Cài đặt",
		"screen_small_warn":    "⚠️  CẢNH BÁO KÍCH THƯỚC MÀN HÌNH QUÁ NHỎ\n\nKích thước hiện tại: %dx%d\nYêu cầu tối thiểu:    80x24\n\nVui lòng phóng to cửa sổ terminal để tiếp tục sử dụng ứng dụng mà không vỡ giao diện.",

		// Status Bar
		"status_home":    "Trang chủ • Nhấn '/' tìm kiếm • 'g' Thể loại • 'H' Lịch sử • 'c' Cài đặt • 'Enter' xem",
		"status_search":  "Tìm kiếm • Nhấn 'Enter' xem kết quả • 'Esc' quay lại",
		"status_detail":  "Chi tiết phim • Dùng ↑/↓/←/→ chọn tập • 'Enter' chọn server phát • 'Esc' quay lại",
		"status_servers": "Nguồn phát • 'Enter' phát video • 'o' mở trình duyệt • 'm' thử MPV • 'Esc' quay lại",
		"status_genre":   "Duyệt theo Thể loại • 'Tab' hoặc ←/→ đổi thể loại • 'Enter' xem chi tiết • 'Esc' quay lại",
		"status_history": "Lịch sử xem phim • 'Enter' tiếp tục xem • 'd' xóa lịch sử • 'Esc' quay lại",
		"status_config":  "Cài đặt cấu hình • ↑/↓ chọn mục • Enter hoặc ←/→ thay đổi • 'Esc' quay lại",

		// Messages
		"mpv_not_installed":  "Chưa cài đặt mpv! Vui lòng làm theo hướng dẫn.",
		"playing_title":      "Đang phát: %s",
		"playback_finished":  "Đã phát xong video.",
		"opened_browser_msg": "Đã mở liên kết nguồn phát trên trình duyệt web mặc định.",
		"mpv_stream_err":     "MPV không thể mở nguồn web này (HTTP 403 / Unsupported). Nhấn 'o' để mở trên trình duyệt web.",

		// Keys Help
		"key_up":       "lên",
		"key_down":     "xuống",
		"key_left":     "trái / tab trước",
		"key_right":    "phải / tab sau",
		"key_tab":      "chuyển tab/nhóm",
		"key_shifttab": "lùi tab/nhóm",
		"key_enter":    "chọn / xem",
		"key_back":     "quay lại",
		"key_search":   "tìm kiếm",
		"key_reload":   "tải lại",
		"key_help":     "trợ giúp",
		"key_quit":     "thoát",
		"key_browser":  "mở trình duyệt",
		"key_config":   "cài đặt",
		"key_genre":    "thể loại",
		"key_history":  "lịch sử xem",

		// Home Screen
		"home_loading":    "Đang tải dữ liệu '%s'...",
		"home_error":      "Lỗi tải trang chủ: %v\nNhấn 'r' để thử lại.",
		"home_empty":      "Không có phim nào trong danh mục này.",
		"tab_anime-new":   "Mới cập nhật",
		"tab_anime-season": "Phim mùa",
		"tab_anime-series": "Anime bộ",
		"tab_anime-single": "Anime lẻ",
		"tab_hh-trungquoc": "Hoạt hình TQ",
		"tab_hot-viewed-today":  "Top ngày",
		"tab_hot-viewed-season": "Top mùa",
		"tab_hot-top-voted":     "Yêu thích",
		"tab_hot-viewed-month":  "Top tháng",
		"tab_top-bo-week":       "Top bộ tuần",
		"tab_top-le-week":       "Top lẻ tuần",

		// Search Screen
		"search_header":      "🔍 TÌM KIẾM ANIME",
		"search_placeholder": "Nhập tên anime cần tìm kiếm (gõ ít nhất 2 ký tự)...",
		"search_loading":    "Đang tìm kiếm...",
		"search_error":      "Lỗi tìm kiếm: %v\nNhấn 'r' để thử lại hoặc 'esc' để đổi từ khóa.",
		"search_sugg_title": "Gợi ý nhanh (nhấn Enter để tìm kiếm đầy đủ):",
		"search_empty":      "Không tìm thấy anime nào phù hợp. Nhấn 'esc' để nhập từ khóa khác.",
		"search_page_info":  "Kết quả cho '%s' - Trang %d/%d (n: tiếp, p: trước, esc: sửa từ khóa)",

		// Detail Screen
		"detail_loading":     "Đang tải thông tin '%s'...",
		"detail_error":       "Lỗi tải chi tiết: %v\nNhấn 'r' để thử lại hoặc 'esc' để quay lại.",
		"detail_no_data":     "Không có dữ liệu phim.",
		"detail_episodes":    "Tổng số tập: %d",
		"detail_desc":        "Mô tả: %s",
		"detail_ep_list":     "DANH SÁCH TẬP PHIM (dùng ↑/↓/←/→ để chọn, Enter để phát):",
		"detail_ep_empty":    "Phim chưa cập nhật tập nào hoặc là phim sắp chiếu.",

		// Servers Screen
		"servers_loading":    "Đang tải danh sách server phát...",
		"servers_resolving":  "Đang phân giải link nguồn phát video...",
		"servers_select":     "CHỌN SERVER PHÁT (dùng ↑/↓ để chọn, Enter để tiếp tục):",
		"servers_default":    "(Mặc định)",
		"servers_quality":    "CHỌN CHẤT LƯỢNG VIDEO (dùng ↑/↓ để chọn, Enter để phát):",
		"servers_source":     "NGUỒN PHÁT [%s]:",
		"servers_play_enter": "👉 Nhấn Enter : Phát video ngay (Mở Cửa sổ Web Player mượt mà)",
		"servers_play_mpv":   "👉 Nhấn Enter : Phát video bằng MPV",
		"servers_open_brow":  "• Nhấn 'o'   : Mở liên kết trên trình duyệt web mặc định",
		"servers_try_mpv":    "• Nhấn 'm'   : Thử mở bằng MPV (nguồn web shield có thể bị 403)",
		"servers_back":       "• Nhấn 'esc' : Quay lại danh sách chọn server khác",
		"servers_auto_open":  "✅ Đã tự động phát video qua %s!",
		"servers_launching":  "🎬 Đang phát video qua %s! (Bạn có thể xem ngay trên desktop)",
		"servers_opened_br":  "🌐 Đã mở liên kết trên trình duyệt web mặc định.",

		// Genre Screen
		"genre_loading":   "Đang tải anime thể loại '%s'...",
		"genre_error":     "Lỗi tải thể loại: %v\nNhấn 'r' để thử lại.",
		"genre_empty":     "Không có phim nào trong thể loại này.",
		"genre_page_info": "Thể loại: %s • Trang %d/%d (n: trang sau, p: trang trước)",

		// History Screen
		"history_header":     "🕒 LỊCH SỬ XEM PHIM",
		"history_empty":      "Chưa có lịch sử xem phim nào. Khi bạn xem phim, lịch sử sẽ tự động lưu tại đây.",
		"history_sub_header": "Tổng cộng: %d phim đã xem • Nhấn 'Enter' để xem tiếp • Nhấn 'd' để xóa lịch sử",
		"time_just_now":      "Vừa xong",
		"time_mins_ago":      "%d phút trước",
		"time_hours_ago":     "%d giờ trước",

		// Config Menu
		"cfg_header":       "⚙️  CÀI ĐẶT CẤU HÌNH (CONFIG MENU)",
		"cfg_lang":         "Ngôn ngữ hiển thị (Language)",
		"cfg_lang_val":     "Tiếng Việt",
		"cfg_player_mode":   "Chế độ phát Video (Player Mode)",
		"cfg_pm_auto":       "Tự động (Cửa sổ App cho Iframe, MPV cho Direct)",
		"cfg_pm_app":        "Cửa sổ Standalone Web Player (Chromium App)",
		"cfg_pm_browser":    "Trình duyệt mặc định (Default Browser)",
		"cfg_pm_mpv":        "MPV Player (Chỉ hỗ trợ Direct Stream)",
		"cfg_auto_play":     "Tự động phát ngay khi chọn tập",
		"cfg_auto_play_on":  "BẬT (Phát ngay)",
		"cfg_auto_play_off": "TẮT (Hỏi trước)",
		"cfg_quality":       "Chất lượng video ưu tiên (cho direct API)",
		"cfg_theme":         "Giao diện Theme",
		"cfg_theme_term":    "Tự động nhận theo Terminal (Kitty/Foot/Alacritty)",
		"cfg_theme_dark":    "Dark Mode",
		"cfg_theme_light":   "Light Mode",
		"cfg_clear_history": "Xóa toàn bộ lịch sử xem phim",
		"cfg_clear_btn":     "[Nhấn Enter để xóa]",
		"cfg_footer":        "• Dùng ↑/↓ để chọn mục • Enter hoặc ←/→ để thay đổi giá trị • Esc để quay lại",
		"cfg_msg_lang":      "Đã đổi ngôn ngữ sang Tiếng Việt.",
		"cfg_msg_pm_auto":   "Chế độ phát: Tự động tối ưu (Cửa sổ App cho Iframe, MPV cho Direct).",
		"cfg_msg_pm_app":    "Chế độ phát: Cửa sổ Standalone Web Player độc lập.",
		"cfg_msg_pm_browser":"Chế độ phát: Trình duyệt mặc định (Default Browser).",
		"cfg_msg_pm_mpv":    "Chế độ phát: Luôn thử mở bằng MPV.",
		"cfg_msg_ap_on":     "Đã bật: Tự động phát ngay khi resolve nguồn.",
		"cfg_msg_ap_off":    "Đã tắt: Hiển thị bảng chọn trước khi phát.",
		"cfg_msg_quality":   "Ưu tiên chất lượng: %s",
		"cfg_msg_theme":     "Theme mode: %s",
		"cfg_confirm_clear": "⚠️ Nhấn Enter lần nữa để xác nhận xóa sạch lịch sử!",
		"cfg_msg_cleared":   "Đã xóa toàn bộ lịch sử xem phim!",
	},

	"en": {
		// App Chrome & Navigation
		"app_title":            "📺 ANIMEVIETSUB TUI",
		"top_search_hint":      "Press '/' for quick search...",
		"top_shortcut_search":  "[/] Search",
		"top_shortcut_genre":   "[g] Genres",
		"top_shortcut_history": "[H] History",
		"top_shortcut_config":  "[c] Settings",
		"screen_small_warn":    "⚠️  SCREEN SIZE TOO SMALL\n\nCurrent dimensions: %dx%d\nMinimum required:  80x24\n\nPlease enlarge your terminal window to continue using the application.",

		// Status Bar
		"status_home":    "Home • Press '/' to search • 'g' Genres • 'H' History • 'c' Settings • 'Enter' Watch",
		"status_search":  "Search • Press 'Enter' for results • 'Esc' Back",
		"status_detail":  "Anime Detail • Use ↑/↓/←/→ to select episode • 'Enter' Select server • 'Esc' Back",
		"status_servers": "Playback • 'Enter' Play video • 'o' Open browser • 'm' Try MPV • 'Esc' Back",
		"status_genre":   "Genres • 'Tab' or ←/→ switch genre • 'Enter' Details • 'Esc' Back",
		"status_history": "Watch History • 'Enter' Continue watching • 'd' Clear history • 'Esc' Back",
		"status_config":  "Settings • ↑/↓ Select • Enter or ←/→ Change • 'Esc' Back",

		// Messages
		"mpv_not_installed":  "MPV is not installed! Please follow the installation guide.",
		"playing_title":      "Playing: %s",
		"playback_finished":  "Playback finished.",
		"opened_browser_msg": "Opened video stream link in default web browser.",
		"mpv_stream_err":     "MPV cannot play this web source (HTTP 403 / Unsupported). Press 'o' to open in browser.",

		// Keys Help
		"key_up":       "up",
		"key_down":     "down",
		"key_left":     "left / prev tab",
		"key_right":    "right / next tab",
		"key_tab":      "switch tab/group",
		"key_shifttab": "prev tab/group",
		"key_enter":    "select / watch",
		"key_back":     "back",
		"key_search":   "search",
		"key_reload":   "reload",
		"key_help":     "help",
		"key_quit":     "quit",
		"key_browser":  "open browser",
		"key_config":   "settings",
		"key_genre":    "genres",
		"key_history":  "watch history",

		// Home Screen
		"home_loading":    "Loading '%s'...",
		"home_error":      "Failed to load home: %v\nPress 'r' to retry.",
		"home_empty":      "No anime found in this category.",
		"tab_anime-new":   "New Releases",
		"tab_anime-season": "Current Season",
		"tab_anime-series": "TV Series",
		"tab_anime-single": "Movies / OVA",
		"tab_hh-trungquoc": "Donghua (3D)",
		"tab_hot-viewed-today":  "Top Today",
		"tab_hot-viewed-season": "Top Season",
		"tab_hot-top-voted":     "Most Favorited",
		"tab_hot-viewed-month":  "Top Month",
		"tab_top-bo-week":       "Top Series (Week)",
		"tab_top-le-week":       "Top Movie (Week)",

		// Search Screen
		"search_header":      "🔍 SEARCH ANIME",
		"search_placeholder": "Enter anime title to search (min 2 chars)...",
		"search_loading":    "Searching...",
		"search_error":      "Search error: %v\nPress 'r' to retry or 'esc' to edit keyword.",
		"search_sugg_title": "Quick suggestions (press Enter for full search):",
		"search_empty":      "No matching anime found. Press 'esc' to try another keyword.",
		"search_page_info":  "Results for '%s' - Page %d/%d (n: next, p: prev, esc: edit keyword)",

		// Detail Screen
		"detail_loading":     "Loading '%s' details...",
		"detail_error":       "Failed to load details: %v\nPress 'r' to retry or 'esc' to go back.",
		"detail_no_data":     "No anime information available.",
		"detail_episodes":    "Episodes: %d",
		"detail_desc":        "Description: %s",
		"detail_ep_list":     "EPISODE LIST (use ↑/↓/←/→ to select, Enter to play):",
		"detail_ep_empty":    "No episodes released yet or upcoming anime.",

		// Servers Screen
		"servers_loading":    "Loading servers...",
		"servers_resolving":  "Resolving video stream link...",
		"servers_select":     "SELECT SERVER (use ↑/↓ to select, Enter to continue):",
		"servers_default":    "(Default)",
		"servers_quality":    "SELECT VIDEO QUALITY (use ↑/↓ to select, Enter to play):",
		"servers_source":     "STREAM SOURCE [%s]:",
		"servers_play_enter": "👉 Press Enter : Play video now (Smooth Standalone Player)",
		"servers_play_mpv":   "👉 Press Enter : Play video using MPV",
		"servers_open_brow":  "• Press 'o'   : Open link in default web browser",
		"servers_try_mpv":    "• Press 'm'   : Try MPV (web shield sources may return 403)",
		"servers_back":       "• Press 'esc' : Go back to server selection",
		"servers_auto_open":  "✅ Automatically started playback via %s!",
		"servers_launching":  "🎬 Playing video via %s! (You can watch now on desktop)",
		"servers_opened_br":  "🌐 Opened video link in default web browser.",

		// Genre Screen
		"genre_loading":   "Loading genre '%s'...",
		"genre_error":     "Failed to load genre: %v\nPress 'r' to retry.",
		"genre_empty":     "No anime found in this genre.",
		"genre_page_info": "Genre: %s • Page %d/%d (n: next, p: prev)",

		// History Screen
		"history_header":     "🕒 WATCH HISTORY",
		"history_empty":      "No watch history yet. When you watch anime, history will be saved here.",
		"history_sub_header": "Total: %d anime watched • Press 'Enter' to resume • Press 'd' to clear history",
		"time_just_now":      "Just now",
		"time_mins_ago":      "%d mins ago",
		"time_hours_ago":     "%d hours ago",

		// Config Menu
		"cfg_header":       "⚙️  SETTINGS (CONFIG MENU)",
		"cfg_lang":         "Language",
		"cfg_lang_val":     "English",
		"cfg_player_mode":   "Player Mode",
		"cfg_pm_auto":       "Auto (App Window for Iframe, MPV for Direct)",
		"cfg_pm_app":        "Standalone Web Player Window (Chromium App)",
		"cfg_pm_browser":    "Default Browser",
		"cfg_pm_mpv":        "MPV Player (Direct Stream only)",
		"cfg_auto_play":     "Auto Play on Selection",
		"cfg_auto_play_on":  "ON (Auto Play)",
		"cfg_auto_play_off": "OFF (Ask First)",
		"cfg_quality":       "Preferred Quality (direct API)",
		"cfg_theme":         "Theme Mode",
		"cfg_theme_term":    "Terminal Adaptive (Kitty/Foot/Alacritty)",
		"cfg_theme_dark":    "Dark Mode",
		"cfg_theme_light":   "Light Mode",
		"cfg_clear_history": "Clear Watch History",
		"cfg_clear_btn":     "[Press Enter to clear]",
		"cfg_footer":        "• Use ↑/↓ to select • Enter or ←/→ to change • Esc to go back",
		"cfg_msg_lang":      "Language switched to English.",
		"cfg_msg_pm_auto":   "Player mode: Auto optimal (App Window for Iframe, MPV for Direct).",
		"cfg_msg_pm_app":    "Player mode: Standalone Web Player Window.",
		"cfg_msg_pm_browser":"Player mode: Default Browser.",
		"cfg_msg_pm_mpv":    "Player mode: Always attempt MPV.",
		"cfg_msg_ap_on":     "Enabled: Automatically start playback upon resolve.",
		"cfg_msg_ap_off":    "Disabled: Show playback selection first.",
		"cfg_msg_quality":   "Preferred quality: %s",
		"cfg_msg_theme":     "Theme mode: %s",
		"cfg_confirm_clear": "⚠️ Press Enter again to confirm clearing all history!",
		"cfg_msg_cleared":   "All watch history cleared!",
	},
}

// T trả về chuỗi văn bản đã dịch theo ngôn ngữ lang ("vi" hoặc "en").
// Nếu không tìm thấy trong lang yêu cầu, tự động fallback sang "vi".
func T(lang, key string, args ...any) string {
	if lang != "en" {
		lang = "vi"
	}

	dict, ok := translations[lang]
	if !ok {
		dict = translations["vi"]
	}

	str, ok := dict[key]
	if !ok {
		// Fallback sang tiếng Việt
		if viDict, viOk := translations["vi"]; viOk {
			str = viDict[key]
		}
	}

	if str == "" {
		str = key
	}

	if len(args) > 0 {
		return fmt.Sprintf(str, args...)
	}
	return str
}

// TranslateTab dịch nhãn của Tab trang chủ.
func TranslateTab(lang, tabType, defaultLabel string) string {
	k := "tab_" + tabType
	translated := T(lang, k)
	if translated != k {
		return translated
	}
	return defaultLabel
}

// FormatTimeAgo định dạng thời gian thân thiện theo ngôn ngữ.
func FormatTimeAgo(lang string, t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return T(lang, "time_just_now")
	} else if diff < time.Hour {
		return T(lang, "time_mins_ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return T(lang, "time_hours_ago", int(diff.Hours()))
	}
	return t.Format("02/01/2006")
}
