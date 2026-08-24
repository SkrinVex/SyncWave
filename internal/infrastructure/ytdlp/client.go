package ytdlp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	ytdlpPath   string
	ffmpegPath  string
	cookiesPath string
	musicDir    string
	coversDir   string
	mu          sync.RWMutex
	proxyURL    string
	audioFormat string
}

type PlaylistEntry struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Uploader   string      `json:"uploader"`
	Artist     string      `json:"artist"`
	Track      string      `json:"track"`
	Channel    string      `json:"channel,omitempty"`
	Creator    string      `json:"creator,omitempty"`
	AltTitle   string      `json:"alt_title,omitempty"`
	Duration   interface{} `json:"duration,omitempty"`
	URL        string      `json:"url"`
	WebpageURL string      `json:"webpage_url"`
}

var (
	progressRegex       = regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+(?:~?\s*)?(\S+)\s+at\s+(\S+)\s+ETA\s+(\S+)`)
	progressSimpleRegex = regexp.MustCompile(`\[download\]\s+([\d\.]+)%`)
	junkRegex           = regexp.MustCompile(`(?i)\s*[\(\[](?:official\s*(?:audio|video|music\s*video|hd|hq|lyric\s*video|mv)?|audio\s*only|audio|lyrics|lyric\s*video|hd|hq|4k|remastered|visualizer)[\)\]]`)
)

func CleanArtist(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "NA" || raw == "Unknown Artist" {
		return ""
	}
	raw = strings.TrimSuffix(raw, " - Topic")
	raw = strings.TrimSuffix(raw, " – Topic")
	return strings.TrimSpace(raw)
}

func CleanTitle(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "NA" || raw == "Unknown Title" {
		return ""
	}
	raw = junkRegex.ReplaceAllString(raw, "")
	return strings.TrimSpace(raw)
}

func (e *PlaylistEntry) GetID() string {
	if e.ID != "" {
		return e.ID
	}
	u := e.URL
	if u == "" {
		u = e.WebpageURL
	}
	if u != "" && strings.Contains(u, "v=") {
		parts := strings.Split(u, "v=")
		if len(parts) > 1 {
			id := parts[1]
			if idx := strings.Index(id, "&"); idx != -1 {
				id = id[:idx]
			}
			return id
		}
	}
	return ""
}

func (e *PlaylistEntry) GetArtist() string {
	if a := CleanArtist(e.Artist); a != "" {
		return a
	}
	if a := CleanArtist(e.Creator); a != "" {
		return a
	}
	if a := CleanArtist(e.Uploader); a != "" {
		return a
	}
	if a := CleanArtist(e.Channel); a != "" {
		return a
	}
	if strings.Contains(e.Title, " - ") {
		parts := strings.SplitN(e.Title, " - ", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			return CleanArtist(parts[0])
		}
	}
	return ""
}

func (e *PlaylistEntry) GetCleanTitle() string {
	title := e.Title
	if e.Track != "" && e.Track != "NA" {
		return CleanTitle(e.Track)
	}
	if strings.Contains(title, " - ") {
		parts := strings.SplitN(title, " - ", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			title = parts[1]
		}
	}
	clean := CleanTitle(title)
	if clean != "" {
		return clean
	}
	return e.Title
}

func (e *PlaylistEntry) GetDuration() int {
	if e.Duration == nil {
		return 0
	}
	switch v := e.Duration.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	}
	return 0
}

type FlatPlaylistOutput struct {
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Entries []PlaylistEntry `json:"entries"`
}

type ProgressCallback func(percent float64, speed string, eta string, status string)

func NormalizePlaylistURL(input string) string {
	input = strings.TrimSpace(input)
	if input == "LM" || input == "liked" {
		return "LM"
	}
	if strings.Contains(input, "list=") {
		u, err := url.Parse(input)
		if err == nil {
			if listID := u.Query().Get("list"); listID != "" {
				return listID
			}
		}
	}
	return input
}

func NewClient(ytdlpPath, ffmpegPath, cookiesPath, musicDir, coversDir string) *Client {
	return &Client{
		ytdlpPath:   ytdlpPath,
		ffmpegPath:  ffmpegPath,
		cookiesPath: cookiesPath,
		musicDir:    musicDir,
		coversDir:   coversDir,
		audioFormat: "opus",
	}
}

func (c *Client) SetProxy(proxyURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.proxyURL = proxyURL
}

func (c *Client) SetAudioFormat(format string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.audioFormat = format
}

func (c *Client) GetMusicDir() string {
	return c.musicDir
}

func (c *Client) GetYTDLPVersion() (string, error) {
	cmd := exec.Command(c.ytdlpPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "unknown", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) GetFFmpegVersion() (string, error) {
	bin := c.ffmpegPath
	if bin == "" {
		bin = "ffmpeg"
	}
	cmd := exec.Command(bin, "-version")
	out, err := cmd.Output()
	if err != nil {
		return "unknown", err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "unknown", nil
}

func (c *Client) GetUserCookiesPath(userID string) string {
	if userID != "" {
		userPath := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
		if info, err := os.Stat(userPath); err == nil && info.Size() > 0 {
			return userPath
		}
		return userPath
	}
	return c.cookiesPath
}

func (c *Client) HasUserCookies(userID string) bool {
	if userID == "" {
		info, err := os.Stat(c.cookiesPath)
		return err == nil && info.Size() > 0
	}
	userPath := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
	info, err := os.Stat(userPath)
	return err == nil && info.Size() > 0
}

func (c *Client) GetUserCookiesModTime(userID string) (string, bool) {
	if userID == "" {
		return c.GetCookiesModTime()
	}
	userPath := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
	if info, err := os.Stat(userPath); err == nil && info.Size() > 0 {
		return info.ModTime().Format(time.RFC3339), true
	}
	return "", false
}

func (c *Client) SaveUserCookies(userID string, content []byte) error {
	normalized := NormalizeCookiesToNetscape(content)
	var path string
	if userID != "" {
		path = filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
	} else {
		path = c.cookiesPath
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, normalized, 0600)
}

func (c *Client) DeleteUserCookies(userID string) error {
	if userID != "" {
		path := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
		return os.Remove(path)
	}
	return os.Remove(c.cookiesPath)
}

func (c *Client) HasCookies() bool {
	info, err := os.Stat(c.cookiesPath)
	return err == nil && info.Size() > 0
}

func (c *Client) GetCookiesModTime() (string, bool) {
	info, err := os.Stat(c.cookiesPath)
	if err != nil || info.Size() == 0 {
		return "", false
	}
	return info.ModTime().Format(time.RFC3339), true
}

func (c *Client) SaveCookies(content []byte) error {
	normalized := NormalizeCookiesToNetscape(content)
	_ = os.MkdirAll(filepath.Dir(c.cookiesPath), 0755)
	return os.WriteFile(c.cookiesPath, normalized, 0600)
}

func (c *Client) ValidateUserCookies(userID string) *CookieValidationResult {
	if userID == "" {
		return ValidateCookieFile(c.cookiesPath)
	}
	userPath := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
	if info, err := os.Stat(userPath); err != nil || info.Size() == 0 {
		return &CookieValidationResult{
			HasCookies:  false,
			IsValid:     false,
			Status:      CookieStatusMissing,
			ErrorReason: "Файл cookies не загружен",
		}
	}
	return ValidateCookieFile(userPath)
}

func (c *Client) buildBaseArgsForUser(userID string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var args []string

	if c.ffmpegPath != "" && c.ffmpegPath != "ffmpeg" && filepath.IsAbs(c.ffmpegPath) {
		args = append(args, "--ffmpeg-location", c.ffmpegPath)
	}

	if c.HasUserCookies(userID) {
		args = append(args, "--cookies", c.GetUserCookiesPath(userID))
	}

	if c.proxyURL != "" {
		args = append(args, "--proxy", c.proxyURL)
	}

	args = append(args,
		"-4",
		"--socket-timeout", "15",
		"-R", "2",
		"--no-mtime",
		"--newline",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
		"--extractor-args", "youtube:skip=translated_subs",
	)

	return args
}

func (c *Client) FetchFlatPlaylist(ctx context.Context, urlOrID string) (*FlatPlaylistOutput, error) {
	return c.FetchFlatPlaylistForUser(ctx, urlOrID, "")
}

func (c *Client) ExtractPlaylistDelta(ctx context.Context, urlOrID string) (*FlatPlaylistOutput, error) {
	return c.FetchFlatPlaylistForUser(ctx, urlOrID, "")
}

func (c *Client) ExtractPlaylistDeltaForUser(ctx context.Context, urlOrID string, userID string) (*FlatPlaylistOutput, error) {
	return c.FetchFlatPlaylistForUser(ctx, urlOrID, userID)
}

func (c *Client) FetchFlatPlaylistForUser(ctx context.Context, urlOrID string, userID string) (*FlatPlaylistOutput, error) {
	targetURL := urlOrID
	if !strings.HasPrefix(urlOrID, "http://") && !strings.HasPrefix(urlOrID, "https://") {
		targetURL = fmt.Sprintf("https://www.youtube.com/playlist?list=%s", urlOrID)
	}

	args := c.buildBaseArgsForUser(userID)
	args = append(args,
		"--flat-playlist",
		"--dump-single-json",
		"--ignore-errors",
		targetURL,
	)

	cookiePath := c.GetUserCookiesPath(userID)
	cookieInfo, cErr := os.Stat(cookiePath)
	cookieSize := int64(0)
	if cErr == nil {
		cookieSize = cookieInfo.Size()
	}
	log.Printf("[yt-dlp] Fetching playlist metadata for %s (cookies: %s, size: %d bytes)", targetURL, cookiePath, cookieSize)
	log.Printf("[yt-dlp CMD] %s %s", c.ytdlpPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		log.Printf("[yt-dlp ERROR] FetchFlatPlaylist failed: %v (stderr: %s)", err, strings.TrimSpace(errStr))
		if isAuth, reason := IsYTDLPAuthError(errStr); isAuth {
			return nil, fmt.Errorf("ошибка авторизации: %s", reason)
		}
		return nil, fmt.Errorf("yt-dlp flat playlist error: %w (%s)", err, strings.TrimSpace(errStr))
	}

	var output FlatPlaylistOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		log.Printf("[yt-dlp ERROR] Failed to parse playlist json: %v", err)
		return nil, fmt.Errorf("failed to parse playlist json: %w", err)
	}

	log.Printf("[yt-dlp SUCCESS] Playlist extracted: %s (found %d entries)", output.Title, len(output.Entries))
	return &output, nil
}

type DownloadResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Duration  int    `json:"duration"`
	FilePath  string `json:"file_path"`
	CoverPath string `json:"cover_path"`
	FileSize  int64  `json:"file_size"`
	Format    string `json:"format"`
	Bitrate   int    `json:"bitrate"`
}

func (c *Client) buildDownloadArgsForUser(targetURL, outTemplate, format string, userID string, withCookies bool) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var args []string

	if c.ffmpegPath != "" && c.ffmpegPath != "ffmpeg" && filepath.IsAbs(c.ffmpegPath) {
		args = append(args, "--ffmpeg-location", c.ffmpegPath)
	}

	if withCookies && c.HasUserCookies(userID) {
		args = append(args, "--cookies", c.GetUserCookiesPath(userID))
	}

	if c.proxyURL != "" {
		args = append(args, "--proxy", c.proxyURL)
	}

	args = append(args,
		"-4",
		"--socket-timeout", "15",
		"-R", "2",
		"--no-mtime",
		"--newline",
		"--progress",
		"--ignore-errors",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
		"--extractor-args", "youtube:skip=translated_subs",
		"-f", "bestaudio/best",
		"--extract-audio",
		"--audio-format", format,
		"--audio-quality", "0",
		"--write-thumbnail",
		"--convert-thumbnails", "jpg",
		"--embed-metadata",
		"--embed-thumbnail",
		"--parse-metadata", "%(release_year,upload_date)s:%(meta_date)s",
		"--parse-metadata", "%(album,playlist,title)s:%(meta_album)s",
		"--parse-metadata", "%(album,title)s:%(meta_album)s",
		"--parse-metadata", "%(track_number,playlist_index)d:%(meta_track)s",
		"--no-playlist",
		"-o", outTemplate,
		"--print", "METADATA:%(id)s|||%(title)s|||%(artist)s|||%(album)s|||%(duration)s|||%(uploader)s|||%(channel)s|||%(track)s|||%(creator)s",
		"--no-simulate",
		targetURL,
	)

	return args
}

func (c *Client) DownloadTrack(ctx context.Context, youtubeID string, onProgress ProgressCallback) (*DownloadResult, error) {
	return c.DownloadTrackForUser(ctx, youtubeID, "", "", "", onProgress)
}

func (c *Client) DownloadTrackForUser(ctx context.Context, youtubeID, title, artist, userID string, onProgress ProgressCallback) (*DownloadResult, error) {
	c.mu.RLock()
	format := c.audioFormat
	if format == "" {
		format = "opus"
	}
	c.mu.RUnlock()

	targetURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)
	outTemplate := filepath.Join(c.musicDir, fmt.Sprintf("%s.%%(ext)s", youtubeID))

	// Track-level timeout: max 3 minutes per single track
	trackCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	hasCookies := c.HasUserCookies(userID)

	// Attempt 1: Direct YouTube URL with cookies (if available)
	log.Printf("[Worker] [Attempt 1/3] Downloading %s (%s - %s) with cookies (hasCookies: %v)", youtubeID, artist, title, hasCookies)
	res, err := c.executeDownloadForUser(trackCtx, youtubeID, targetURL, outTemplate, format, userID, hasCookies, onProgress)
	if err == nil {
		return res, nil
	}
	log.Printf("[Worker] [Attempt 1/3] Failed for %s: %v", youtubeID, err)

	// Attempt 2: If failed with cookies, retry without cookies
	if hasCookies && trackCtx.Err() == nil {
		log.Printf("[Worker] [Attempt 2/3] Retrying %s (%s - %s) WITHOUT cookies", youtubeID, artist, title)
		res, err2 := c.executeDownloadForUser(trackCtx, youtubeID, targetURL, outTemplate, format, userID, false, onProgress)
		if err2 == nil {
			return res, nil
		}
		log.Printf("[Worker] [Attempt 2/3] Failed for %s: %v", youtubeID, err2)
		err = err2
	}

	// Attempt 3: If direct video ID failed (e.g. removed/unavailable/blocked) but title is known, search alternative
	if trackCtx.Err() == nil && title != "" && title != youtubeID {
		cleanA := CleanArtist(artist)
		cleanT := CleanTitle(title)
		if cleanT == "" {
			cleanT = title
		}
		searchQuery := cleanT
		if cleanA != "" && !strings.Contains(strings.ToLower(cleanT), strings.ToLower(cleanA)) {
			searchQuery = fmt.Sprintf("%s %s", cleanA, cleanT)
		}
		searchURL := fmt.Sprintf("ytsearch1:%s", searchQuery)

		if hasCookies {
			log.Printf("[Worker] [Attempt 3/3-A] Searching alternative for %s: '%s' with cookies", youtubeID, searchQuery)
			res3, err3 := c.executeDownloadForUser(trackCtx, youtubeID, searchURL, outTemplate, format, userID, true, onProgress)
			if err3 == nil {
				res3.ID = youtubeID
				if res3.Title == "" || res3.Title == "Unknown Title" {
					res3.Title = cleanT
				}
				if res3.Artist == "" || res3.Artist == "Unknown Artist" {
					res3.Artist = cleanA
				}
				return res3, nil
			}
			log.Printf("[Worker] [Attempt 3/3-A] Search with cookies failed for %s: %v", youtubeID, err3)
		}

		log.Printf("[Worker] [Attempt 3/3-B] Searching alternative for %s: '%s' WITHOUT cookies", youtubeID, searchQuery)
		res4, err4 := c.executeDownloadForUser(trackCtx, youtubeID, searchURL, outTemplate, format, userID, false, onProgress)
		if err4 == nil {
			res4.ID = youtubeID
			if res4.Title == "" || res4.Title == "Unknown Title" {
				res4.Title = cleanT
			}
			if res4.Artist == "" || res4.Artist == "Unknown Artist" {
				res4.Artist = cleanA
			}
			return res4, nil
		}
		log.Printf("[Worker] [Attempt 3/3-B] Search without cookies failed for %s: %v", youtubeID, err4)
	}

	return nil, err
}

func (c *Client) executeDownloadForUser(ctx context.Context, targetTrackID, targetURL, outTemplate, format string, userID string, withCookies bool, onProgress ProgressCallback) (*DownloadResult, error) {
	args := c.buildDownloadArgsForUser(targetURL, outTemplate, format, userID, withCookies)

	cookiePath := c.GetUserCookiesPath(userID)
	cookieInfo, cErr := os.Stat(cookiePath)
	cookieSize := int64(0)
	if cErr == nil {
		cookieSize = cookieInfo.Size()
	}

	log.Printf("[yt-dlp] Downloading track %s (URL: %s, withCookies: %v, cookies: %s, size: %d bytes)", targetTrackID, targetURL, withCookies, cookiePath, cookieSize)
	log.Printf("[yt-dlp CMD] %s %s", c.ytdlpPath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	var metadataLine string
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	parseProgressLine := func(line string) {
		if onProgress != nil {
			if matches := progressRegex.FindStringSubmatch(line); len(matches) >= 5 {
				pct, _ := strconv.ParseFloat(matches[1], 64)
				onProgress(pct, matches[3], matches[4], "downloading")
			} else if matches := progressSimpleRegex.FindStringSubmatch(line); len(matches) >= 2 {
				pct, _ := strconv.ParseFloat(matches[1], 64)
				onProgress(pct, "", "", "downloading")
			} else if strings.Contains(line, "[ExtractAudio]") || strings.Contains(line, "[ThumbnailsConvertor]") || strings.Contains(line, "[EmbedThumbnail]") || strings.Contains(line, "[Metadata]") {
				onProgress(100, "", "", "processing")
			}
		}
	}

	// Stream stdout in real-time
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[yt-dlp stdout] %s", line)

			if strings.HasPrefix(line, "METADATA:") {
				mu.Lock()
				metadataLine = line
				mu.Unlock()
			}

			parseProgressLine(line)
		}
	}()

	// Stream stderr in real-time
	var stderrBuf bytes.Buffer
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[yt-dlp stderr] %s", line)
			mu.Lock()
			stderrBuf.WriteString(line)
			stderrBuf.WriteString("\n")
			mu.Unlock()

			parseProgressLine(line)
		}
	}()

	cmdErr := cmd.Wait()
	wg.Wait()

	mu.Lock()
	stdErrStr := strings.TrimSpace(stderrBuf.String())
	mu.Unlock()

	log.Printf("[yt-dlp EXIT] track=%s, exitErr=%v, stderrLen=%d", targetTrackID, cmdErr, len(stdErrStr))

	youtubeID := targetTrackID
	if youtubeID == "" {
		if idx := strings.Index(targetURL, "v="); idx != -1 {
			youtubeID = targetURL[idx+2:]
			if amp := strings.Index(youtubeID, "&"); amp != -1 {
				youtubeID = youtubeID[:amp]
			}
		}
	}

	expectedAudioPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.%s", youtubeID, format))
	fi, err := os.Stat(expectedAudioPath)
	if err != nil {
		matches, globErr := filepath.Glob(filepath.Join(c.musicDir, fmt.Sprintf("%s.*", youtubeID)))
		if globErr == nil && len(matches) > 0 {
			for _, m := range matches {
				if !strings.HasSuffix(m, ".jpg") && !strings.HasSuffix(m, ".webp") && !strings.HasSuffix(m, ".png") && !strings.HasSuffix(m, ".part") && !strings.HasSuffix(m, ".temp") && !strings.HasSuffix(m, ".ytdl") {
					expectedAudioPath = m
					fi, _ = os.Stat(expectedAudioPath)
					break
				}
			}
		}
	}

	// If audio file was successfully written to disk (>=300KB), we recover from non-fatal exit errors (such as ffprobe thumbnail warnings)
	if fi != nil && fi.Size() >= 300*1024 {
		if cmdErr != nil {
			log.Printf("[yt-dlp RECOVERED] Valid audio file exists on disk (%s, %d bytes) despite exit code: %v", expectedAudioPath, fi.Size(), cmdErr)
			cmdErr = nil
		}
	} else if fi != nil && fi.Size() < 300*1024 && cmdErr != nil {
		// Incomplete / corrupted stub file produced on error -> delete it so fallback search can run!
		log.Printf("[yt-dlp WARN] Deleting incomplete/corrupted file stub (%s, %d bytes)", expectedAudioPath, fi.Size())
		_ = os.Remove(expectedAudioPath)
		fi = nil
	}

	if cmdErr != nil {
		if isAuth, reason := IsYTDLPAuthError(stdErrStr); isAuth {
			return nil, fmt.Errorf("Ошибка авторизации YouTube (%s): %s", reason, stdErrStr)
		}
		return nil, fmt.Errorf("download failed: %w (stderr: %s)", cmdErr, stdErrStr)
	}

	// Optimize .m4a / .mp4 containers with faststart for instant HTTP audio playback
	if fi != nil && (strings.HasSuffix(expectedAudioPath, ".m4a") || strings.HasSuffix(expectedAudioPath, ".mp4")) {
		ffmpegBin := c.ffmpegPath
		if ffmpegBin == "" {
			ffmpegBin = "ffmpeg"
		}
		tmpFast := expectedAudioPath + ".fast.m4a"
		fastCmd := exec.Command(ffmpegBin, "-y", "-i", expectedAudioPath, "-c", "copy", "-movflags", "+faststart", tmpFast)
		if fErr := fastCmd.Run(); fErr == nil {
			_ = os.Rename(tmpFast, expectedAudioPath)
			if updatedFi, sErr := os.Stat(expectedAudioPath); sErr == nil {
				fi = updatedFi
			}
			log.Printf("[yt-dlp] Applied faststart optimization to %s", expectedAudioPath)
		} else {
			_ = os.Remove(tmpFast)
		}
	}

	var fileSize int64
	if fi != nil {
		fileSize = fi.Size()
	}

	// Detect cover file
	expectedCoverPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.jpg", youtubeID))
	finalCoverPath := filepath.Join(c.coversDir, fmt.Sprintf("%s.jpg", youtubeID))
	if _, err := os.Stat(expectedCoverPath); err == nil {
		_ = os.MkdirAll(c.coversDir, 0755)
		_ = os.Rename(expectedCoverPath, finalCoverPath)
	} else if _, err := os.Stat(finalCoverPath); err == nil {
		// Already in covers dir
	} else {
		// Look for webp or other cover formats in musicDir
		coverMatches, _ := filepath.Glob(filepath.Join(c.musicDir, fmt.Sprintf("%s.*", youtubeID)))
		for _, cm := range coverMatches {
			if strings.HasSuffix(cm, ".jpg") || strings.HasSuffix(cm, ".webp") || strings.HasSuffix(cm, ".png") {
				_ = os.MkdirAll(c.coversDir, 0755)
				_ = os.Rename(cm, finalCoverPath)
				break
			}
		}
	}

	if _, err := os.Stat(finalCoverPath); err != nil {
		finalCoverPath = ""
	}

	res := &DownloadResult{
		ID:        youtubeID,
		Title:     "Unknown Title",
		Artist:    "Unknown Artist",
		Album:     "",
		Duration:  0,
		FilePath:  expectedAudioPath,
		CoverPath: finalCoverPath,
		FileSize:  fileSize,
		Format:    format,
		Bitrate:   160,
	}

	mu.Lock()
	meta := metadataLine
	mu.Unlock()

	if meta != "" {
		parts := strings.Split(strings.TrimPrefix(meta, "METADATA:"), "|||")
		if len(parts) >= 1 && parts[0] != "" {
			res.ID = youtubeID
		}
		if len(parts) >= 2 && parts[1] != "" && parts[1] != "NA" {
			res.Title = CleanTitle(parts[1])
		}
		if len(parts) >= 8 && parts[7] != "" && parts[7] != "NA" {
			res.Title = CleanTitle(parts[7])
		}

		if len(parts) >= 3 && parts[2] != "" && parts[2] != "NA" {
			res.Artist = CleanArtist(parts[2])
		} else if len(parts) >= 9 && parts[8] != "" && parts[8] != "NA" {
			res.Artist = CleanArtist(parts[8])
		} else if len(parts) >= 6 && parts[5] != "" && parts[5] != "NA" {
			res.Artist = CleanArtist(parts[5])
		} else if len(parts) >= 7 && parts[6] != "" && parts[6] != "NA" {
			res.Artist = CleanArtist(parts[6])
		}

		if len(parts) >= 4 && parts[3] != "" && parts[3] != "NA" {
			res.Album = strings.TrimSpace(parts[3])
		}
		if len(parts) >= 5 && parts[4] != "" && parts[4] != "NA" {
			sec, _ := strconv.Atoi(parts[4])
			res.Duration = sec
		}
	}

	// Smart parse Title & Artist if Artist is NA or missing
	if res.Artist == "" || res.Artist == "NA" || res.Artist == "Unknown Artist" {
		if strings.Contains(res.Title, " - ") {
			subParts := strings.SplitN(res.Title, " - ", 2)
			res.Artist = CleanArtist(subParts[0])
			res.Title = CleanTitle(subParts[1])
		}
	}
	if res.Artist == "" {
		res.Artist = "Unknown Artist"
	}
	if res.Title == "" {
		res.Title = "Unknown Title"
	}

	return res, nil
}

func (c *Client) TestProxyConnection(ctx context.Context, proxyURLStr string) error {
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return fmt.Errorf("invalid proxy URL format: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   8 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.youtube.com/generate_204", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("proxy test failed (cannot reach YouTube): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("proxy responded with error status: %s", resp.Status)
	}

	return nil
}

// OptimizeLibraryFaststart scans music directory and ensures all .m4a / .mp4 files have the moov atom at the front for instant streaming
func (c *Client) OptimizeLibraryFaststart() {
	if c.musicDir == "" {
		return
	}
	log.Printf("[Worker] Starting Faststart optimization scan in %s...", c.musicDir)
	var allAudio []string
	if matches, err := filepath.Glob(filepath.Join(c.musicDir, "*.m4a")); err == nil {
		allAudio = append(allAudio, matches...)
	}
	if matches, err := filepath.Glob(filepath.Join(c.musicDir, "*.mp4")); err == nil {
		allAudio = append(allAudio, matches...)
	}

	if len(allAudio) == 0 {
		log.Printf("[Worker] Faststart optimization: No .m4a/.mp4 files found in %s", c.musicDir)
		return
	}

	ffmpegBin := c.ffmpegPath
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}

	optimizedCount := 0
	for _, m := range allAudio {
		// Skip files smaller than 100KB
		if fi, err := os.Stat(m); err == nil && fi.Size() < 100*1024 {
			continue
		}
		tmpFast := m + ".fast.m4a"
		fastCmd := exec.Command(ffmpegBin, "-y", "-i", m, "-c", "copy", "-movflags", "+faststart", tmpFast)
		if fErr := fastCmd.Run(); fErr == nil {
			_ = os.Rename(tmpFast, m)
			optimizedCount++
		} else {
			_ = os.Remove(tmpFast)
		}
	}
	log.Printf("[Worker] Faststart optimization verified and applied for %d/%d audio files", optimizedCount, len(allAudio))
}
