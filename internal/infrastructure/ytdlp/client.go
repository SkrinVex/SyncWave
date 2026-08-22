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
	Duration   interface{} `json:"duration,omitempty"`
	URL        string      `json:"url"`
	WebpageURL string      `json:"webpage_url"`
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

func (c *Client) buildBaseArgs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var args []string

	if c.ffmpegPath != "" && c.ffmpegPath != "ffmpeg" && filepath.IsAbs(c.ffmpegPath) {
		args = append(args, "--ffmpeg-location", c.ffmpegPath)
	}

	if c.HasCookies() {
		args = append(args, "--cookies", c.cookiesPath)
		args = append(args, "--extractor-args", "youtube:player_skip=configs")
	}

	if c.proxyURL != "" {
		args = append(args, "--proxy", c.proxyURL)
	}

	args = append(args,
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"--no-check-certificates",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
	)

	return args
}

// NormalizePlaylistURL converts "LM", playlist IDs, or full URLs into standard yt-dlp target URLs
func NormalizePlaylistURL(rawInput string) string {
	raw := strings.TrimSpace(rawInput)
	if raw == "LM" || raw == "liked" || raw == "likes" {
		return "https://music.youtube.com/playlist?list=LM"
	}
	if raw == "LL" {
		return "https://www.youtube.com/playlist?list=LL"
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		if strings.Contains(raw, "list=") {
			parts := strings.Split(raw, "list=")
			if len(parts) > 1 {
				playlistID := parts[1]
				if idx := strings.Index(playlistID, "&"); idx != -1 {
					playlistID = playlistID[:idx]
				}
				if idx := strings.Index(playlistID, "#"); idx != -1 {
					playlistID = playlistID[:idx]
				}
				if strings.Contains(raw, "music.youtube.com") {
					return "https://music.youtube.com/playlist?list=" + playlistID
				}
				return "https://www.youtube.com/playlist?list=" + playlistID
			}
		}
		return raw
	}

	if strings.HasPrefix(raw, "PL") || strings.HasPrefix(raw, "RD") || strings.HasPrefix(raw, "OL") || strings.HasPrefix(raw, "VL") || strings.HasPrefix(raw, "LRSR") {
		return "https://music.youtube.com/playlist?list=" + raw
	}

	return "https://music.youtube.com/playlist?list=" + raw
}

// ExtractPlaylistDelta fetches fast JSON metadata (--flat-playlist) without downloading media streams
func (c *Client) ExtractPlaylistDelta(ctx context.Context, playlistInput string) (*FlatPlaylistOutput, error) {
	url := NormalizePlaylistURL(playlistInput)

	out, err := c.runFlatPlaylist(ctx, url)
	if err == nil && len(out.Entries) > 0 {
		return out, nil
	}

	// Fallback 1: If URL was music.youtube.com, try www.youtube.com
	if strings.Contains(url, "music.youtube.com/playlist?list=") {
		ytUrl := strings.Replace(url, "music.youtube.com", "www.youtube.com", 1)
		if ytOut, ytErr := c.runFlatPlaylist(ctx, ytUrl); ytErr == nil && len(ytOut.Entries) > 0 {
			return ytOut, nil
		}
	}

	// Fallback 2: If playlist was LM (Liked Music), also try LL (Liked Videos)
	if strings.Contains(url, "list=LM") {
		llUrl := "https://www.youtube.com/playlist?list=LL"
		if llOut, llErr := c.runFlatPlaylist(ctx, llUrl); llErr == nil && len(llOut.Entries) > 0 {
			return llOut, nil
		}
	}

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) runFlatPlaylist(ctx context.Context, targetURL string) (*FlatPlaylistOutput, error) {
	args := append(c.buildBaseArgs(),
		"--flat-playlist",
		"-J",
		"--yes-playlist",
		"--ignore-errors",
		"--no-warnings",
		targetURL,
	)

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract playlist: %v, stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	var output FlatPlaylistOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err == nil && len(output.Entries) > 0 {
		return &output, nil
	}

	// Dynamic fallback map unmarshaling for custom formats
	var rawMap map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &rawMap); err == nil {
		var entries []PlaylistEntry
		if rawEntries, ok := rawMap["entries"].([]interface{}); ok {
			for _, item := range rawEntries {
				if itemMap, ok := item.(map[string]interface{}); ok {
					var e PlaylistEntry
					if id, ok := itemMap["id"].(string); ok {
						e.ID = id
					}
					if title, ok := itemMap["title"].(string); ok {
						e.Title = title
					}
					if uploader, ok := itemMap["uploader"].(string); ok {
						e.Uploader = uploader
					}
					if artist, ok := itemMap["artist"].(string); ok {
						e.Artist = artist
					}
					if u, ok := itemMap["url"].(string); ok {
						e.URL = u
					}
					if dur, ok := itemMap["duration"]; ok {
						e.Duration = dur
					}
					if e.GetID() != "" {
						entries = append(entries, e)
					}
				}
			}
		}
		title, _ := rawMap["title"].(string)
		id, _ := rawMap["id"].(string)
		return &FlatPlaylistOutput{
			ID:      id,
			Title:   title,
			Entries: entries,
		}, nil
	}

	return nil, fmt.Errorf("failed to parse playlist json")
}

type DownloadResult struct {
	YouTubeID string
	Title     string
	Artist    string
	Album     string
	Duration  int
	FilePath  string
	CoverPath string
	FileSize  int64
	Format    string
	Bitrate   int
}

var (
	progressRegex       = regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+(?:~?\s*)?(\S+)\s+at\s+(\S+)\s+ETA\s+(\S+)`)
	progressSimpleRegex = regexp.MustCompile(`\[download\]\s+([\d\.]+)%`)
)

func (c *Client) buildDownloadArgs(targetURL, outTemplate, format string, withCookies bool) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var args []string

	if c.ffmpegPath != "" && c.ffmpegPath != "ffmpeg" && filepath.IsAbs(c.ffmpegPath) {
		args = append(args, "--ffmpeg-location", c.ffmpegPath)
	}

	if withCookies && c.HasCookies() {
		args = append(args, "--cookies", c.cookiesPath)
	}

	if c.proxyURL != "" {
		args = append(args, "--proxy", c.proxyURL)
	}

	if withCookies {
		args = append(args, "--extractor-args", "youtube:player_skip=configs")
	}

	args = append(args,
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"--no-check-certificates",
		"--no-warnings",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
		"-f", "bestaudio/best",
		"-x",
		"--audio-format", format,
		"--audio-quality", "0",
		"--embed-thumbnail",
		"--add-metadata",
		"--write-thumbnail",
		"--convert-thumbnails", "jpg",
		"--no-playlist",
		"-o", outTemplate,
		"--print", "METADATA:%(id)s|||%(title)s|||%(artist)s|||%(album)s|||%(duration)s|||%(uploader)s|||%(channel)s",
		"--no-simulate",
		"--no-quiet",
		"--newline",
		targetURL,
	)

	return args
}

// DownloadTrack downloads a single track by YouTube ID or URL, extracts audio, tags metadata and creates cover image
func (c *Client) DownloadTrack(ctx context.Context, youtubeID string, onProgress ProgressCallback) (*DownloadResult, error) {
	c.mu.RLock()
	format := c.audioFormat
	if format == "" {
		format = "opus"
	}
	c.mu.RUnlock()

	targetURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", youtubeID)
	outTemplate := filepath.Join(c.musicDir, fmt.Sprintf("%s.%%(ext)s", youtubeID))

	// Track-level timeout: max 2 minutes per single track
	trackCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// First attempt: Android player client without cookies (fastest and most reliable)
	res, err := c.executeDownload(trackCtx, youtubeID, targetURL, outTemplate, format, false, onProgress)
	if err == nil {
		return res, nil
	}

	// Second attempt (if cookies exist): Retry with cookies enabled if first attempt failed
	if c.HasCookies() && ctx.Err() == nil {
		return c.executeDownload(trackCtx, youtubeID, targetURL, outTemplate, format, true, onProgress)
	}

	return nil, err
}

func (c *Client) executeDownload(ctx context.Context, youtubeID, targetURL, outTemplate, format string, withCookies bool, onProgress ProgressCallback) (*DownloadResult, error) {
	args := c.buildDownloadArgs(targetURL, outTemplate, format, withCookies)

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp download: %w", err)
	}

	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	title := youtubeID
	artist := "Unknown Artist"
	album := ""
	duration := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Parse custom lightweight metadata
		if strings.HasPrefix(line, "METADATA:") {
			raw := strings.TrimPrefix(line, "METADATA:")
			parts := strings.Split(raw, "|||")
			if len(parts) >= 2 && parts[1] != "NA" && parts[1] != "" {
				title = parts[1]
			}
			if len(parts) >= 3 && parts[2] != "NA" && parts[2] != "" {
				artist = parts[2]
			}
			if len(parts) >= 4 && parts[3] != "NA" && parts[3] != "" {
				album = parts[3]
			}
			if len(parts) >= 5 && parts[4] != "NA" && parts[4] != "" {
				if d, err := strconv.Atoi(parts[4]); err == nil {
					duration = d
				}
			}
			if artist == "Unknown Artist" || artist == "" {
				if len(parts) >= 6 && parts[5] != "NA" && parts[5] != "" {
					artist = parts[5]
				} else if len(parts) >= 7 && parts[6] != "NA" && parts[6] != "" {
					artist = parts[6]
				}
			}
			continue
		}

		// Parse download progress
		if onProgress != nil {
			if matches := progressRegex.FindStringSubmatch(line); len(matches) >= 5 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				speed := matches[3]
				eta := matches[4]
				onProgress(percent, speed, eta, "downloading")
			} else if matches := progressSimpleRegex.FindStringSubmatch(line); len(matches) >= 2 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				onProgress(percent, "", "", "downloading")
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v, stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	// Locate downloaded audio file
	expectedAudioPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.%s", youtubeID, format))
	fileInfo, err := os.Stat(expectedAudioPath)
	if err != nil {
		pattern := filepath.Join(c.musicDir, fmt.Sprintf("%s.*", youtubeID))
		matches, _ := filepath.Glob(pattern)
		found := false
		for _, m := range matches {
			ext := strings.ToLower(filepath.Ext(m))
			if ext == ".opus" || ext == ".m4a" || ext == ".mp3" || ext == ".flac" || ext == ".webm" || ext == ".ogg" {
				expectedAudioPath = m
				fileInfo, _ = os.Stat(m)
				format = strings.TrimPrefix(ext, ".")
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("could not locate downloaded audio file for %s", youtubeID)
		}
	}

	// Handle thumbnail / cover art
	expectedCoverPath := filepath.Join(c.coversDir, fmt.Sprintf("%s.jpg", youtubeID))
	tempThumbnailPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.jpg", youtubeID))
	if _, err := os.Stat(tempThumbnailPath); err == nil {
		_ = os.Rename(tempThumbnailPath, expectedCoverPath)
	} else {
		webpPattern := filepath.Join(c.musicDir, fmt.Sprintf("%s.webp", youtubeID))
		if _, err := os.Stat(webpPattern); err == nil {
			_ = os.Rename(webpPattern, expectedCoverPath)
		}
	}

	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	return &DownloadResult{
		YouTubeID: youtubeID,
		Title:     title,
		Artist:    artist,
		Album:     album,
		Duration:  duration,
		FilePath:  expectedAudioPath,
		CoverPath: expectedCoverPath,
		FileSize:  fileSize,
		Format:    format,
		Bitrate:   160,
	}, nil
}

func (c *Client) TestProxyConnection(ctx context.Context, proxyURL string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	if proxyURL != "" {
		pURL, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy url: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(pURL),
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.youtube.com/generate_204", nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("proxy connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
