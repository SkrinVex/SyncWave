# Спецификация REST API и Стриминга SyncWave

SyncWave предоставляет производительный REST API с авторизацией по JWT, поддержкой стриминга по стандарту RFC 7233 (HTTP 206 Partial Content) для веб- и мобильных плееров, а также потоком Server-Sent Events (SSE) для трансляции прогресса скачивания в реальном времени.

---

## Аутентификация

Все защищенные эндпоинты принимают JWT-токен одним из двух способов:

1. **HTTP-заголовок (Стандартный REST)**:
   ```http
   Authorization: Bearer <jwt_токен>
   ```

2. **Параметр в URL `?token=` (Для `<audio>`, `<img>` и ExoPlayer)**:
   ```http
   GET /api/v1/tracks/d1c258ec-7e5c-4433-a3d5-dcb348987ec8/stream?token=<jwt_токен>
   ```

---

## Список эндпоинтов

### 1. Авторизация и первичная настройка

#### `GET /api/v1/auth/status`
Проверяет, требуется ли первичная инициализация администратора (первый запуск).

**Ответ `200 OK`**:
```json
{
  "needs_setup": false
}
```

#### `POST /api/v1/auth/setup`
Создает учетную запись администратора при первом запуске.

**Тело запроса**:
```json
{
  "username": "admin",
  "password": "your-password"
}
```

#### `POST /api/v1/auth/login`
Авторизация пользователя. Возвращает JWT-токен со сроком действия 30 дней.

#### `GET /api/v1/auth/me` *(Protected)*
Возвращает профиль текущего пользователя.

---

### 2. Музыкальная библиотека и треки

#### `GET /api/v1/tracks` *(Protected)*
Получение списка треков с поиском, фильтрацией и пагинацией.

**Query-параметры**:
* `q` (string) — поиск по названию, исполнителю или альбому;
* `playlist_id` (string) — фильтр по ID плейлиста;
* `status` (string) — `ready`, `downloading`, `failed`;
* `sort_by` (string) — `created_at`, `title`, `artist`, `duration`;
* `order` (string) — `asc` или `desc` (по умолчанию `desc`);
* `page` (int) — номер страницы (по умолчанию `1`);
* `page_size` (int) — размер страницы (по умолчанию `50`).

**Пример ответа `200 OK`**:
```json
{
  "tracks": [
    {
      "id": "d1c258ec-7e5c-4433-a3d5-dcb348987ec8",
      "youtube_id": "dQw4w9WgXcQ",
      "playlist_id": "99ea7021-39bc-4a41-a67b-12d8376eef5a",
      "title": "Never Gonna Give You Up",
      "artist": "Rick Astley",
      "album": "Whenever You Need Somebody",
      "duration": 213,
      "file_size": 4325810,
      "format": "opus",
      "bitrate": 160,
      "status": "ready",
      "downloaded_at": "2026-08-23T00:15:00Z",
      "created_at": "2026-08-23T00:15:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 50,
  "total_pages": 1
}
```

#### `GET /api/v1/tracks/stats` *(Protected)*
Возвращает сводную статистику библиотеки: общее количество треков, готовых к воспроизведению, общий объем файлов в байтах и суммарную длительность аудио.

#### `GET /api/v1/tracks/{id}` *(Protected)*
Информация о конкретном треке.

#### `DELETE /api/v1/tracks/{id}` *(Protected)*
Удаляет запись из БД и физически удаляет аудиофайл и обложку с диска `/data`.

---

### 3. Стриминг аудио и раздача медиа

#### `GET /api/v1/tracks/{id}/stream` *(Protected)*
Стриминг аудиофайла с поддержкой заголовка `Range: bytes=...` и ответа `206 Partial Content`. Позволяет плеерам мгновенно перематывать трек на любую секунду без скачивания всего файла.

**Пример запроса с перемоткой**:
```http
GET /api/v1/tracks/d1c258ec-7e5c-4433-a3d5-dcb348987ec8/stream HTTP/1.1
Host: localhost:8080
Range: bytes=1048576-2097152
```

**Ответ `206 Partial Content`**:
```http
HTTP/1.1 206 Partial Content
Content-Type: audio/ogg; codecs=opus
Content-Range: bytes 1048576-2097152/4325810
Content-Length: 1048577
Accept-Ranges: bytes
Cache-Control: public, max-age=3600
```

#### `GET /api/v1/tracks/{id}/cover` *(Protected)*
Раздача обложки трека в формате JPEG с кэширующими заголовками (`Cache-Control: public, max-age=31536000, immutable`).

#### `GET /api/v1/tracks/{id}/download` *(Protected)*
Прямое скачивание аудиофайла с именем `Исполнитель - Название.opus` через заголовок `Content-Disposition: attachment`.

---

### 4. Управление плейлистами

#### `GET /api/v1/playlists` *(Protected)*
Список всех плейлистов пользователя с подсчетом количества треков.

#### `POST /api/v1/playlists` *(Protected)*
Добавление нового плейлиста для автосинхронизации.

**Тело запроса**:
```json
{
  "title": "Понравившиеся",
  "url_or_id": "LM",
  "auto_sync": true,
  "sync_interval_minutes": 60
}
```

#### `PUT /api/v1/playlists/{id}` *(Protected)*
Обновление названия, интервала синхронизации или переключателя автосинхронизации.

#### `DELETE /api/v1/playlists/{id}` *(Protected)*
Удаление подписки на плейлист (уже скачанные треки сохраняются в библиотеке).

#### `POST /api/v1/playlists/{id}/sync` *(Protected)*
Ручной запуск синхронизации конкретного плейлиста.

---

### 5. Фоновый воркер и SSE-телеметрия

#### `POST /api/v1/sync/trigger` *(Protected)*
Запуск синхронизации всех плейлистов.

#### `GET /api/v1/sync/progress` *(Protected)*
Текущее состояние фонового воркера (активен ли, процент выполнения, текущий трек, скорость, ETA).

#### `GET /api/v1/sync/events` *(Protected)*
Поток Server-Sent Events (SSE) для мгновенного обновления прогресса и логов в веб-интерфейсе:
```
event: message
data: {"type":"progress","data":{"active":true,"current_track_title":"Solaris","percentage":68.4,"speed":"2.8MiB/s","eta":"00:01"}}

event: message
data: {"type":"log","data":{"id":42,"level":"success","message":"Successfully archived: Photay - Solaris","created_at":"2026-08-23T00:20:00Z"}}
```

#### `GET /api/v1/sync/logs` *(Protected)*
Получение последних записей логов воркера (`?limit=100`).

---

### 6. Настройки и диагностика

#### `GET /api/v1/settings` *(Protected)*
Системные настройки (прокси, кодек, размер базы данных, занятое место на диске, версии yt-dlp и ffmpeg).

#### `PUT /api/v1/settings` *(Protected)*
Обновление параметров (прокси, кодек, лимит воркеров).

#### `POST /api/v1/settings/cookies` *(Protected)*
Загрузка файла `cookies.txt` (`multipart/form-data` или текстовое тело).

#### `POST /api/v1/settings/test-proxy` *(Protected)*
Проверка работоспособности HTTP/SOCKS5 прокси.

