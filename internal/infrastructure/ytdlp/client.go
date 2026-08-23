package ytdlp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	}
	if info, err := os.Stat(c.cookiesPath); err == nil && info.Size() > 0 {
		return c.cookiesPath
	}
	return c.cookiesPath
}

func (c *Client) HasUserCookies(userID string) bool {
	path := c.GetUserCookiesPath(userID)
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return true
	}
	return false
}

func (c *Client) GetUserCookiesModTime(userID string) (string, bool) {
	path := c.GetUserCookiesPath(userID)
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return info.ModTime().Format(time.RFC3339), true
	}
	return "", false
}

func (c *Client) SaveUserCookies(userID string, content []byte) error {
	var path string
	if userID != "" {
		path = filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
	} else {
		path = c.cookiesPath
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, content, 0600)
}

func (c *Client) DeleteUserCookies(userID string) error {
	if userID != "" {
		path := filepath.Join(filepath.Dir(c.cookiesPath), "cookies", fmt.Sprintf("cookies_%s.txt", userID))
		_ = os.Remove(path)
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
	_ = os.MkdirAll(filepath.Dir(c.cookiesPath), 0755)
	return os.WriteFile(c.cookiesPath, content, 0600)
}

func (c *Client) ValidateUserCookies(userID string) *CookieValidationResult {
	path := c.GetUserCookiesPath(userID)
	return ValidateCookieFile(path)
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
		"--no-check-certificates",
		"--no-warnings",
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
		"--no-warnings",
		"--ignore-errors",
		targetURL,
	)

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if isAuth, reason := IsYTDLPAuthError(errStr); isAuth {
			return nil, fmt.Errorf("ошибка авторизации: %s", reason)
		}
		return nil, fmt.Errorf("yt-dlp flat playlist error: %w (%s)", err, strings.TrimSpace(errStr))
	}

	var output FlatPlaylistOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse playlist json: %w", err)
	}

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
		"--no-check-certificates",
		"--no-warnings",
		"--extractor-args", "youtube:skip=translated_subs",
		"-f", "bestaudio/best",
		"--extract-audio",
		"--audio-format", format,
		"--audio-quality", "0",
		"--postprocessor-args", "ffmpeg:-movflags +faststart",
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
		"--no-quiet",
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
	res, err := c.executeDownloadForUser(trackCtx, youtubeID, targetURL, outTemplate, format, userID, hasCookies, onProgress)
	if err == nil {
		return res, nil
	}

	// Attempt 2: If failed with cookies, retry without cookies
	if hasCookies && trackCtx.Err() == nil {
		res, err2 := c.executeDownloadForUser(trackCtx, youtubeID, targetURL, outTemplate, format, userID, false, onProgress)
		if err2 == nil {
			return res, nil
		}
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
		}

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
	}

	return nil, err
}

func (c *Client) executeDownloadForUser(ctx context.Context, targetTrackID, targetURL, outTemplate, format string, userID string, withCookies bool, onProgress ProgressCallback) (*DownloadResult, error) {
	args := c.buildDownloadArgsForUser(targetURL, outTemplate, format, userID, withCookies)

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

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "METADATA:") {
				mu.Lock()
				metadataLine = line
				mu.Unlock()
			}

			if onProgress != nil {
				if matches := progressRegex.FindStringSubmatch(line); len(matches) >= 5 {
					pct, _ := strconv.ParseFloat(matches[1], 64)
					onProgress(pct, matches[3], matches[4], "downloading")
				} else if matches := progressSimpleRegex.FindStringSubmatch(line); len(matches) >= 2 {
					pct, _ := strconv.ParseFloat(matches[1], 64)
					onProgress(pct, "", "", "downloading")
				}
			}
		}
	}()

	var stderrBuf bytes.Buffer
	go func() {
		_, _ = stderrBuf.ReadFrom(stderrPipe)
	}()

	cmdErr := cmd.Wait()
	if cmdErr != nil {
		stdErrStr := stderrBuf.String()
		if isAuth, reason := IsYTDLPAuthError(stdErrStr); isAuth {
			return nil, fmt.Errorf("Ошибка авторизации YouTube (%s): %s", reason, stdErrStr)
		}
		return nil, fmt.Errorf("download failed: %w (stderr: %s)", cmdErr, stdErrStr)
	}

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
				if !strings.HasSuffix(m, ".jpg") && !strings.HasSuffix(m, ".webp") && !strings.HasSuffix(m, ".png") {
					expectedAudioPath = m
					fi, _ = os.Stat(expectedAudioPath)
					break
				}
			}
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
