# Tài Liệu Kỹ Thuật Hệ Thống API AnimeVietSub (animevietsub.nl)

Tài liệu này tổng hợp toàn bộ các API endpoints, cấu trúc dữ liệu và phương thức giao tiếp của website `https://animevietsub.nl/` được trích xuất và phân tích từ network traffic và client script.

---

## 1. Tổng Quan Kiến Trúc Hệ Thống

- **Base URL:** `https://animevietsub.nl` (hoặc `https://www.animevietsub.nl`)
- **Kiến trúc dữ liệu:**
  - **Server-Side Rendering (SSR):** Dành cho trang chủ, trang chi tiết phim, trang danh mục và trang kết quả tìm kiếm (phục vụ SEO).
  - **AJAX / RPC Endpoints:** Phần lớn nằm dưới tiền tố `/ajax/...`, phục vụ các tính năng động (phát video, gợi ý tìm kiếm, lọc danh mục, bình luận, tương tác người dùng).
- **Player CDN & Video Hosting:**
  - CDN Player chính: `https://storage.googleapiscdn.com`
  - Host bên thứ ba: `Abyssplayer`, `Google Drive`, `YouTube`
- **Realtime Service:** Centrifugo v5 WebSocket tại `wss://rt.animevietsub.nl/connection/websocket`.

---

## 2. Bảng Tổng Hợp Endpoints

| Chức năng | Endpoint | Phương thức | Định dạng trả về | Tham số chính |
| :--- | :--- | :---: | :---: | :--- |
| **Gợi ý tìm kiếm** | `/ajax/suggest` | `POST` | HTML snippet | `ajaxSearch=1, keysearch={kw}` |
| **Tìm kiếm phim** | `/tim-kiem/{kw}/` | `GET` | HTML document | Phân trang: `/trang-{page}.html` |
| **Lọc Tab phim trang chủ** | `/ajax/item` | `POST` | HTML snippet | `widget=list-film, type={tag}` |
| **Lấy danh sách Server backup** | `/ajax/player` | `POST` | JSON | `episodeId={epId}, backup=1` |
| **Lấy link Player / Stream** | `/ajax/player` | `POST` | JSON | `link={href}, play={tech}, id={svId}, backuplinks=1` |
| **Chuyển tập phim (hash)** | `/ajax/player` | `POST` | JSON | `link={hash}, id={epId}` |
| **Reload danh sách Server** | `/ajax/player` | `POST` | JSON | `filmLoadserver=1, filmId={id}` |
| **Đánh giá phim (Rating)** | `/ajax/rate` | `GET` | JSON | `_fxAjax=1, score={1-10}, film={id}` |
| **Kiểm tra Bookmark** | `/ajax/notification` | `GET` | JSON | `Bookmark=true, filmId={id}` |
| **Thêm Bookmark** | `/ajax/suggest` | `GET` | JSON | `Bookmark=true, filmId={id}, type=add` |
| **Xóa Bookmark** | `/ajax/notification` | `GET` | JSON | `Bookmark=true, filmId={id}, type=remove` |
| **User Hover Card** | `/ajax/user-card` | `GET` | JSON | `uid={userId}` |
| **Danh sách Fan Cứng** | `/ajax/film-fans` | `GET` | JSON | `film_id={id}, page={p}, size={n}` |
| **Danh sách Bình luận** | `/ajax/comment` | `GET` | JSON | `action=get, film_id={id}, sort={sort}, offset={n}` |
| **Bình luận gần đây / Top** | `/ajax/comment` | `GET` | JSON | `action=get_recent` hoặc `action=get_top_comments` |
| **Đăng bình luận / Phản hồi** | `/ajax/comment` | `POST` | JSON | `action=post, film_id, episode_id, content` |
| **Vote bình luận** | `/ajax/comment` | `POST` | JSON | `action=vote, comment_id, vote_type` |
| **Token Realtime WebSocket** | `/ajax/comment` | `GET` | JSON | `action=token` |
| **Báo lỗi tập phim** | `/ajax/report-error` | `POST` | JSON | `film_id, episode_id, error_type, error_description` |
| **Thông báo tập phim** | `/ajax/all` | `POST` | Text/HTML | `EpisodeMess=1, EpisodeID={epId}` |

---

## 3. Chi Tiết Kỹ Thuật Các API

### 3.1. Nhóm API Player & Streaming Video

#### A. Dữ liệu Player mặc định tại Trang Xem Phim
Khi người dùng truy cập trang `https://animevietsub.nl/phim/{slug}/tap-{ep}-{episodeId}.html`, mã HTML nhúng sẵn dữ liệu player trong biến JavaScript:

```javascript
window.PLAYER_DATA = {
  "_fxStatus": 1,
  "success": 1,
  "title": "AnimeVsub",
  "link": "https://storage.googleapiscdn.com/player/{hash}?isFinal=1",
  "playTech": "iframe", // hoặc "api", "embed"
  "img": "https://cdn.animevietsub.nl/data/banner/....webp",
  "episode_id": "116019"
};
window._epHash = "_avEpuHLG6aYePa3LB2ataDciDjcRXV...";
window._epID = 116019;
```

#### B. API Lấy Danh Sách Server Dự Phòng (Backup Servers)
Gọi sau khi trang xem phim tải xong hoặc khi người dùng hover/click khu vực server list.

- **Endpoint:** `POST https://animevietsub.nl/ajax/player`
- **Headers:**
  - `Content-Type: application/x-www-form-urlencoded; charset=UTF-8`
  - `X-Requested-With: XMLHttpRequest`
- **Request Body:**
  ```text
  episodeId={episodeId}&backup=1
  ```
- **Response Format:** `JSON`
- **Response Ví dụ:**
  ```json
  {
    "success": 1,
    "html": "<h3>Chọn Server:</h3> <a class=\"btn3dsv active\" href=\"#\" data-id=\"0\" data-play=\"api\" data-href=\"10ll-9U-af172X7uBJp_O4...\">DU</a><a class=\"btn3dsv\" href=\"#\" data-id=\"3\" data-play=\"embed\" data-href=\"OETdcNKiTBVyIBdvYpq...\">HDX(ADS)</a>"
  }
  ```
  - `data-id`: ID định danh server (`0`: server chính DU, `3`: server HDX, ...).
  - `data-play`: Chế độ hiển thị player (`api`, `embed`, `iframe`).
  - `data-href`: Token mã hóa thông tin nguồn phát của server đó.

#### C. API Lấy Nguồn Phát Từ Server Đã Chọn / Đổi Tập Phim
- **Endpoint:** `POST https://animevietsub.nl/ajax/player`
- **Headers:**
  - `Content-Type: application/x-www-form-urlencoded; charset=UTF-8`
  - `X-Requested-With: XMLHttpRequest`
- **Request Body:**
  - Khi đổi server backup:
    ```text
    link={data-href}&play={data-play}&id={data-id}&backuplinks=1
    ```
  - Khi chuyển tập phim qua hash:
    ```text
    link={_epHash}&id={_epID}
    ```
- **Response Format:** `JSON`
  - **Trường hợp 1 (Server chính DU - iframe player):**
    ```json
    {
      "_fxStatus": 1,
      "success": 1,
      "title": "AnimeVsub",
      "link": "https://storage.googleapiscdn.com/player/8731e6b4498079dd99461b77bc2ad57cba7a5c4893525a93abe600a4b7430b1b",
      "playTech": "iframe"
    }
    ```
  - **Trường hợp 2 (Server Embed bên ngoài như Abyssplayer, Google Drive):**
    ```json
    {
      "_fxStatus": 1,
      "success": 1,
      "title": "AnimeVsub",
      "link": "https://abyssplayer.com/7C099vMDo",
      "playTech": "embed"
    }
    ```
  - **Trường hợp 3 (Direct MP4 / M3U8):**
    ```json
    {
      "_fxStatus": 1,
      "success": 1,
      "title": "AnimeVsub",
      "link": [
        { "file": "https://.../video-720.mp4", "label": "720p", "type": "mp4" },
        { "file": "https://.../video-1080.mp4", "label": "1080p", "type": "mp4" }
      ],
      "playTech": "api"
    }
    ```

#### D. Cơ chế hoạt động của Player Client:
- `playTech == "iframe"`: Nhúng `<iframe src="{link}">` vào container `#media-player`. Player này tải JWPlayer + HLS.js từ CDN `storage.googleapiscdn.com`.
- `playTech == "embed"`: Nhúng iframe của bên thứ ba.
- `playTech == "api"`: Khởi tạo player native (JWPlayer) phát trực tiếp file stream hoặc file MP4.

---

### 3.2. Nhóm API Tìm Kiếm & Lọc Phim

#### A. Gợi ý tìm kiếm nhanh (Search Suggestion)
- **Endpoint:** `POST https://animevietsub.nl/ajax/suggest`
- **Headers:**
  - `Content-Type: application/x-www-form-urlencoded; charset=UTF-8`
  - `X-Requested-With: XMLHttpRequest`
- **Request Body:**
  ```text
  ajaxSearch=1&keysearch={keyword}
  ```
- **Response Format:** `HTML Snippet` (danh sách thẻ `<li>`):
  ```html
  <li>
    <a style="background-image: url('https://cdn.animevietsub.nl/data/poster/...jpg')" class="thumb" href="https://animevietsub.nl/phim/.../"></a>
    <div class="ss-info">
      <a href="https://animevietsub.nl/phim/.../" class="ss-title">Tên Phim</a>
      <p>Full VietSub</p>
    </div>
    <div class="clearfix"></div>
  </li>
  ```

#### B. Lọc danh sách phim trang chủ theo Tab (Home Tab Filter)
- **Endpoint:** `POST https://animevietsub.nl/ajax/item`
- **Headers:**
  - `Content-Type: application/x-www-form-urlencoded; charset=UTF-8`
  - `X-Requested-With: XMLHttpRequest`
- **Request Body:**
  ```text
  widget=list-film&type={type}
  ```
- **Các giá trị của tham số `type`:**
  - `anime-new`: Phim mới cập nhật
  - `anime-season`: Phim theo mùa hiện tại
  - `anime-series`: Anime bộ (TV Series)
  - `anime-single`: Anime lẻ / Movie / OVA
  - `hh-trungquoc`: Hoạt hình Trung Quốc (3D)
  - `hot-viewed-today`: Xem nhiều nhất hôm nay
  - `hot-viewed-season`: Xem nhiều nhất mùa này
  - `hot-top-voted`: Phim được yêu thích nhất
  - `hot-viewed-month`: Xem nhiều trong tháng
  - `top-bo-week`: Top anime bộ tuần
  - `top-le-week`: Top anime lẻ tuần
- **Response Format:** `HTML Snippet` gồm các thẻ `<li class="TPostMv">...</li>`.

---

### 3.3. Nhóm Dữ Liệu Phim & Tập Phim (SSR Data)

Trên trang chi tiết phim `https://animevietsub.nl/phim/{slug}-a{filmId}/`:

1. **Biến JavaScript toàn cục:**
   ```javascript
   filmInfo.filmID = parseInt('6076');
   filmInfo.title = "Tên Anime";
   filmInfo.fullUrl = "https://animevietsub.nl/phim/.../";
   ```
2. **Metadata cấu trúc (JSON-LD):** Thẻ `<script type="application/ld+json">` chứa thông tin Schema `TVSeries` hoặc `Movie`.
3. **Danh sách tập phim:** Thẻ `<ul class="list-episode">`:
   ```html
   <li class="episode">
     <a href="https://animevietsub.nl/phim/.../tap-01-116019.html"
        title="Tập 01"
        class="btn-episode episode-link"
        data-id="116019"
        data-hash="_avEpuHLG6aYePa3LB2ataDciDjcRXV...">01</a>
   </li>
   ```

---

### 3.4. Hệ Thống API Bình Luận (Comments API)

Tất cả các hành động bình luận đều giao tiếp qua endpoint: `https://animevietsub.nl/ajax/comment`

#### A. Lấy danh sách bình luận của phim / tập
- **Method:** `GET`
- **Query Parameters:**
  - `action`: `get`
  - `film_id`: ID phim
  - `sort`: `newest` hoặc `top`
  - `offset`: Vị trí phân trang (bắt đầu từ `0`)
  - `page_url`: URL trang hiện tại (nếu ở trang xem tập phim)
- **Response Format:** `JSON`
  ```json
  {
    "success": true,
    "total": 24,
    "has_more": false,
    "comments": [
      {
        "id": 105432,
        "film_id": 6076,
        "episode_id": 116019,
        "user_id": 12,
        "user_name": "username",
        "user_fullname": "Tên Hiển Thị",
        "user_avatar": "https://cdn.animevietsub.nl/data/avatar/...jpg",
        "content": "Nội dung bình luận...",
        "votes_up": 5,
        "votes_down": 0,
        "replies_count": 1,
        "is_spoiler": 0,
        "created_at": "2026-10-08 10:00:00"
      }
    ]
  }
  ```

#### B. Lấy phản hồi (Replies) của bình luận
- **Method:** `GET`
- **Query Parameters:**
  - `action`: `get_replies`
  - `parent_id`: ID bình luận cha
  - `offset`: Phân trang replies

#### C. Lấy bình luận mới nhất / Top bình luận toàn trang
- **Method:** `GET`
- **Query Parameters:**
  - `action`: `get_recent` hoặc `get_top_comments`
  - `last_id`: ID bình luận cuối cùng đã nhận
  - `_`: Timestamp chống cache

#### D. Đăng bình luận / Phản hồi mới
- **Method:** `POST`
- **Request Body (Form Data):**
  - `action`: `post`
  - `film_id`: ID phim
  - `episode_id`: ID tập phim
  - `parent_id`: `0` (bình luận gốc) hoặc ID comment cha (phản hồi)
  - `content`: Nội dung bình luận
  - `is_spoiler`: `0` hoặc `1`
  - `thread_key`: Slug phim hoặc URL

#### E. Vote Like / Dislike bình luận
- **Method:** `POST`
- **Request Body:**
  - `action`: `vote`
  - `comment_id`: ID bình luận
  - `vote_type`: `1` (Like) hoặc `-1` (Dislike)
- **Response:** JSON `{ "success": true, "votes_up": 6, "votes_down": 0 }`

#### F. Lấy Token WebSocket Realtime
- **Method:** `GET`
- **Query Parameters:** `action=token`
- **Response:** JSON chứa JWT token để xác thực với Centrifugo WebSocket: `wss://rt.animevietsub.nl/connection/websocket`.

---

### 3.5. Nhóm API Người Dùng, Tương Tác & Đánh Giá

#### A. Đánh giá điểm phim (Rating)
- **Endpoint:** `GET https://animevietsub.nl/ajax/rate`
- **Query Parameters:**
  - `_fxAjax`: `1`
  - `_fxResponseType`: `json`
  - `score`: Điểm đánh giá (1 đến 10)
  - `film`: ID phim
- **Response:**
  ```json
  {
    "_fxStatus": 1,
    "_fxMessage": "Cảm ơn bạn đã đánh giá phim!",
    "rateCount": 769,
    "ratePoint": 6.3
  }
  ```

#### B. Quản lý Yêu thích / Bookmark
- **Kiểm tra trạng thái:** `GET /ajax/notification?Bookmark=true&filmId={filmId}`  
  - Response: `{"status": 1}` (đã theo dõi) hoặc `{"status": 0}`
- **Thêm yêu thích:** `GET /ajax/suggest?Bookmark=true&filmId={filmId}&type=add`  
  - Response: `{"status": 1}`
- **Hủy yêu thích:** `GET /ajax/notification?Bookmark=true&filmId={filmId}&type=remove`  
  - Response: `{"status": 1}`

#### C. User Hover Card (Xem tóm tắt thông tin người dùng)
- **Endpoint:** `GET https://animevietsub.nl/ajax/user-card`
- **Query Parameters:** `uid={userId}`
- **Response:**
  ```json
  {
    "status": 1,
    "data": {
      "user_id": 1,
      "fullname": "Osin",
      "handle": "Osin",
      "avatar": "https://cdn.animevietsub.nl/data/avatar/user-1.gif",
      "bio": "",
      "is_verified": false,
      "joined": "01/2020",
      "gender": { "code": "1", "label": "Nam", "icon": "fa-mars" },
      "profile_url": "/thanh-vien/1/"
    }
  }
  ```

#### D. Danh sách Fan Cứng của phim
- **Endpoint:** `GET https://animevietsub.nl/ajax/film-fans`
- **Query Parameters:**
  - `film_id`: ID phim
  - `page`: Trang số
  - `size`: Kích thước (mặc định 20)
- **Response:**
  ```json
  {
    "success": true,
    "fans": [
      {
        "profile_url": "/thanh-vien/123/",
        "user_avatar": "https://cdn.animevietsub.nl/data/avatar/...",
        "user_name": "Username",
        "user_fullname": "Tên Hiển Thị",
        "valid_count": 45
      }
    ],
    "has_more": false
  }
  ```

#### E. Thông báo người dùng (Notifications)
- **Lấy danh sách thông báo:** `GET /ajax/notification?notif=true`
- **Đánh dấu đã đọc:** `POST /ajax/notification` với body `MarkRead=true&id={notifId}`
- **Xóa thông báo:** `POST /ajax/notification` với body `Delete=true&id={notifId}` (hoặc `Delete=all`)

---

### 3.6. Nhóm API Báo Lỗi & Thông Báo Phụ

#### A. Báo lỗi tập phim
- **Endpoint:** `POST https://animevietsub.nl/ajax/report-error`
- **Headers:** `Content-Type: application/x-www-form-urlencoded; charset=UTF-8`
- **Request Body:**
  - `film_id`: ID phim
  - `episode_id`: ID tập phim
  - `error_type`: Phân loại lỗi (`sub`: phụ đề, `video`: lỗi hình ảnh/tiếng, `watch`: lỗi không xem được, `other`: lỗi khác)
  - `error_description`: Mô tả chi tiết
- **Response Format:** `JSON` `{ "success": true, "message": "Báo lỗi thành công!" }`

#### B. Thông báo riêng của tập phim
- **Endpoint:** `POST https://animevietsub.nl/ajax/all`
- **Request Body:** `EpisodeMess=1&EpisodeID={episodeId}`
- **Response:** `0` nếu không có thông báo, hoặc đoạn mã HTML hiển thị box thông báo.

---

## 4. Các Header Cần Thiết Khi Gửi Request

Để tránh bị chặn hoặc trả về lỗi, các request AJAX cần đính kèm các headers tiêu chuẩn:
```http
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
Referer: https://animevietsub.nl/
X-Requested-With: XMLHttpRequest
Content-Type: application/x-www-form-urlencoded; charset=UTF-8
```
