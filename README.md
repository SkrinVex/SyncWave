# SyncWave 🌊

> **Self-hosted сервис для автоматической синхронизации, резервного копирования и стриминга треков из YouTube Music на ваш личный сервер.**  
> *Полная независимость от облаков и стримингов, высокое качество звука, единый бинарник и стильный веб-интерфейс.*

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Vue 3](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vuedotjs)](https://vuejs.org)
[![SQLite](https://img.shields.io/badge/SQLite-WAL-003B57?style=flat-square&logo=sqlite)](https://sqlite.org)
[![Docker Ready](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![Clean Architecture](https://img.shields.io/badge/Architecture-Clean-indigo?style=flat-square)]()

---

## 🌟 Ключевые возможности

- **🎧 Автоматическая синхронизация и бэкап**: Добавьте любой плейлист YouTube Music или специальный пресет **«Понравившиеся»** (`LM`). Сервис периодически проверяет обновления и скачивает новые треки в фоне.
- **⚡ Алгоритм Delta-синхронизации (Anti-Throttling)**: Сначала запрашиваются только ID треков через `--flat-playlist` без скачивания потоков. Это предотвращает бан по IP от YouTube.
- **🛡️ Защищенная очередь загрузок**: Конкурентность ограничена (1–2 потока) с рандомизированными задержками (jitter), поддержкой сессионных `cookies.txt` и прокси (HTTP/SOCKS5).
- **🔊 Настоящий стриминг (HTTP Range 206)**: Полноценная поддержка частичного контента через `http.ServeContent` — мгновенная перемотка, поддержка HTML5 `<audio>` и кэширования в мобильных плеерах (ExoPlayer/Media3).
- **💎 Минималистичный Studio Web UI**: Никаких шаблонных градиентов — строгий монохромный дизайн **Obsidian Studio** (Vue 3, Tailwind CSS, Pinia) с тактильным плеером, очередью треков, поиском и фильтрами.
- **📡 Телеметрия в реальном времени (SSE)**: Сервер вещает статус скачивания (процент, скорость МБ/с, ETA) и консольные логи прямо в веб-интерфейс через Server-Sent Events.
- **📱 PWA & Экран блокировки (Media Session)**: Полноценная поддержка установки приложения на смартфоны и ПК. Воспроизведение продолжается в фоне, треками можно управлять прямо с экрана блокировки и из системных уведомлений с отображением обложек.
- **📦 Один бинарник со встроенным фронтендом**: Фронтенд встроен в Go-бинарник через `embed.FS`. Для работы нужен лишь один процесс или легковесный Docker-контейнер.
- **📱 Готовность к Android-клиенту**: API полностью оптимизирован под интеграцию с Android Jetpack Compose + Media3 (`SimpleCache` / `CacheDataSource`).

---

## 📐 Архитектура системы

```mermaid
graph TD
    subgraph "Внешние сервисы"
        YTM[YouTube Music API / Web]
        Proxy[HTTP / SOCKS5 Прокси (Опционально)]
    end

    subgraph "SyncWave Server (Go Binary)"
        Router[Chi HTTP Router & SSE Hub]
        AuthMid[JWT Auth Middleware]
        StaticFS[embed.FS Web SPA]
        
        subgraph "Clean Architecture Core"
            Usecases[Track / Playlist / Sync / Settings Usecases]
            WorkerQueue[Очередь воркеров & Cron Планировщик]
            YTDLPWrapper[yt-dlp + FFmpeg Exec Engine]
            SQLiteRepo[SQLite Репозиторий WAL Mode]
        end
    end

    subgraph "Хранилище и Клиенты"
        DBFile[(SQLite: /data/syncwave.db)]
        MusicDir[Музыка и Обложки: /data/music & /data/covers]
        WebUI[Vue 3 Studio Web App]
        Android[Android App / ExoPlayer]
    end

    YTM -->|--flat-playlist delta| YTDLPWrapper
    Proxy -.-> YTDLPWrapper
    YTDLPWrapper -->|Аудио + Метаданные| MusicDir
    WorkerQueue --> YTDLPWrapper
    WorkerQueue --> SQLiteRepo
    SQLiteRepo --> DBFile

    WebUI -->|REST / SSE / Range Stream| Router
    Android -->|REST / Range Stream| Router
    Router --> AuthMid
    Router --> StaticFS
    AuthMid --> Usecases
    Usecases --> SQLiteRepo
    Usecases --> MusicDir
```

---

## 🚀 Быстрый старт через Docker Compose

Самый простой способ развернуть SyncWave на сервере:

1. Создайте файл конфигурации `.env` (или скопируйте его из `.env.example`):
```bash
cp .env.example .env
# Отредактируйте .env, чтобы изменить порт или JWT_SECRET, если нужно
```

2. В `docker-compose.yml` убедитесь, что он выглядит так:
```yaml
# docker-compose.yml
services:
  syncwave:
    build: .
    container_name: syncwave
    restart: unless-stopped
    ports:
      - "${SYNCWAVE_PORT:-8080}:8080"
    env_file:
      - .env
    volumes:
      - ./data:/data
```

Запуск:
```bash
docker compose up -d --build
```

Откройте браузер по адресу **`http://localhost:8080`** и задайте логин/пароль администратора при первом входе.

---

## 🛠️ Локальная сборка без Docker (Single Binary)

### Системные требования:
- **Go 1.24+**
- **Node.js 20+** и **npm**
- Установленные в системе `yt-dlp` и `ffmpeg`

### 1. Сборка фронтенда:
```bash
cd web
npm install
npm run build
cd ..
```

### 2. Сборка Go-бинарника:
```bash
go build -o syncwave cmd/server/main.go
```

### 3. Запуск:
```bash
PORT=8080 DATA_DIR=./data ./syncwave
```

---

## 🍪 Синхронизация «Понравившихся» (Liked Music)

Для доступа к вашему списку «Понравившиеся» и приватным плейлистам YouTube Music требуются куки авторизации:

1. Авторизуйтесь на сайте [music.youtube.com](https://music.youtube.com) в браузере.
2. Экспортируйте файл куков с помощью расширения **Get cookies.txt locally** или **Cookie-Editor**.
3. В веб-интерфейсе SyncWave откройте вкладку **Настройки** (Settings) и перетащите файл `cookies.txt` в окно загрузки.
4. Во вкладке **Плейлисты** нажмите **Добавить плейлист** и выберите пресет **Liked Music (`LM`)**.

📖 Подробная инструкция: [**docs/COOKIES_GUIDE.md**](docs/COOKIES_GUIDE.md)

---

## 🎹 Горячие клавиши веб-плеера

| Клавиша | Действие |
| :--- | :--- |
| <kbd>Space</kbd> | Воспроизведение / Пауза |
| <kbd>←</kbd> / <kbd>→</kbd> | Перемотка назад / вперед на 5 сек |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Громкость выше / ниже на 5% |
| <kbd>M</kbd> | Включить / выключить звук |
| <kbd>L</kbd> | Режим повтора (Выкл / Все / Один трек) |
| <kbd>S</kbd> | Перемешать очередь (Shuffle) |

---

## 📂 Структура проекта (Clean Architecture)

```
SyncWave/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа, DI и graceful shutdown
├── internal/
│   ├── config/                  # Конфигурация из переменных окружения
│   ├── domain/                  # Сущности (Track, Playlist, User), интерфейсы
│   ├── repository/sqlite/       # Репозиторий SQLite (WAL, чистый Go без CGO)
│   ├── usecase/                 # Бизнес-логика (Auth, Track, Playlist, Sync, Settings)
│   ├── infrastructure/
│   │   ├── auth/                # JWT генератор/валидатор, Bcrypt
│   │   ├── worker/              # Очередь загрузок, Cron-планировщик, SSE Hub
│   │   └── ytdlp/               # Обертка yt-dlp, FFmpeg тегирование, обложки
│   └── delivery/http/
│       ├── handler/             # REST & HTTP 206 Range Stream хендлеры
│       ├── middleware/          # JWT Auth, CORS, Logger
│       └── router.go            # Chi роутер
├── web/                         # Vue 3 + Tailwind CSS + Pinia фронтенд
│   ├── src/
│   │   ├── components/          # Плеер, очередь, карточки, скраббер
│   │   ├── stores/              # Pinia хранилища (Player, Tracks, Sync, Auth)
│   │   └── views/               # Библиотека, Плейлисты, Логи, Настройки
│   ├── embed.go                 # Директива go:embed для сборки в один бинарник
│   └── vite.config.js
├── docs/
│   ├── API.md                   # Спецификация REST API и Range Streaming
│   ├── ARCHITECTURE.md          # Подробный разбор архитектуры и решений
│   └── COOKIES_GUIDE.md         # Руководство по экспорту cookies.txt
├── .github/workflows/
│   └── docker-publish.yml       # Автоматическая сборка в GitHub Packages (GHCR)
├── Dockerfile                   # Multi-stage Dockerfile
├── docker-compose.yml           # Готовый compose-файл
└── README.md
```

---

## 📖 Документация

- [**Спецификация REST API и Стриминга**](docs/API.md)
- [**Архитектурное описание проекта**](docs/ARCHITECTURE.md)
- [**Инструкция по экспорту Cookies**](docs/COOKIES_GUIDE.md)

---

## 📄 Лицензия

MIT License © 2026 SyncWave.
