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
	// Fallback to root cookies file if it exists
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
		"--no-check-certificates",
		"--no-warnings",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
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
		targetURL = fmt.Sprintf("https://music.youtube.com/playlist?list=%s", urlOrID)
	}

	args := c.buildBaseArgsForUser(userID)
	args = append(args,
		"--flat-playlist",
		"--dump-single-json",
		"--no-playlist-reverse",
		targetURL,
	)

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stdErrStr := stderr.String()
		if strings.Contains(stdErrStr, "429") {
			return nil, fmt.Errorf("YouTube rate limit (HTTP 429). Please configure residential proxy in Settings: %s", stdErrStr)
		}
		if strings.Contains(stdErrStr, "Private video") || strings.Contains(stdErrStr, "Sign in") {
			return nil, fmt.Errorf("Playlist requires authentication. Please upload cookies.txt in Settings: %s", stdErrStr)
		}
		return nil, fmt.Errorf("yt-dlp flat playlist error: %w (stderr: %s)", err, stdErrStr)
	}

	var flatOutput FlatPlaylistOutput
	if err := json.Unmarshal(stdout.Bytes(), &flatOutput); err != nil {
		return nil, fmt.Errorf("failed to parse playlist json: %w", err)
	}

	return &flatOutput, nil
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

var (
	progressRegex       = regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+(?:~?\s*)?(\S+)\s+at\s+(\S+)\s+ETA\s+(\S+)`)
	progressSimpleRegex = regexp.MustCompile(`\[download\]\s+([\d\.]+)%`)
)

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
		"--newline",
		"--no-check-certificates",
		"--no-warnings",
		"--js-runtimes", "node",
		"--remote-components", "ejs:github",
		"-f", "bestaudio/best",
		"--extract-audio",
		"--audio-format", format,
		"--audio-quality", "0",
		"--embed-thumbnail",
		"--add-metadata",
		"--no-playlist",
		"-o", outTemplate,
		"--print", "METADATA:%(id)s|||%(title)s|||%(artist)s|||%(album)s|||%(duration)s|||%(uploader)s|||%(channel)s",
		"--no-simulate",
		"--no-quiet",
		targetURL,
	)

	return args
}

func (c *Client) DownloadTrack(ctx context.Context, youtubeID string, onProgress ProgressCallback) (*DownloadResult, error) {
	return c.DownloadTrackForUser(ctx, youtubeID, "", onProgress)
}

func (c *Client) DownloadTrackForUser(ctx context.Context, youtubeID string, userID string, onProgress ProgressCallback) (*DownloadResult, error) {
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

	// Attempt 1: Standard YouTube URL with cookies (if available)
	res, err := c.executeDownloadForUser(trackCtx, targetURL, outTemplate, format, userID, hasCookies, onProgress)
	if err == nil {
		return res, nil
	}

	// Attempt 2: If failed with cookies, retry without cookies (fixes expired session / client mismatch errors)
	if hasCookies && trackCtx.Err() == nil {
		res, err2 := c.executeDownloadForUser(trackCtx, targetURL, outTemplate, format, userID, false, onProgress)
		if err2 == nil {
			return res, nil
		}
		err = err2
	}

	return nil, err
}

func (c *Client) executeDownloadForUser(ctx context.Context, targetURL, outTemplate, format string, userID string, withCookies bool, onProgress ProgressCallback) (*DownloadResult, error) {
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
		return nil, fmt.Errorf("download failed: %w (stderr: %s)", cmdErr, stderrBuf.String())
	}

	// Extract YouTube ID from targetURL
	var youtubeID string
	if idx := strings.Index(targetURL, "v="); idx != -1 {
		youtubeID = targetURL[idx+2:]
		if amp := strings.Index(youtubeID, "&"); amp != -1 {
			youtubeID = youtubeID[:amp]
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

	expectedCoverPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.jpg", youtubeID))
	finalCoverPath := filepath.Join(c.coversDir, fmt.Sprintf("%s.jpg", youtubeID))
	if _, err := os.Stat(expectedCoverPath); err == nil {
		_ = os.MkdirAll(c.coversDir, 0755)
		_ = os.Rename(expectedCoverPath, finalCoverPath)
	} else {
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
			res.ID = parts[0]
		}
		if len(parts) >= 2 && parts[1] != "" {
			res.Title = parts[1]
		}
		if len(parts) >= 3 && parts[2] != "" {
			res.Artist = parts[2]
		} else if len(parts) >= 6 && parts[5] != "" {
			res.Artist = parts[5]
		} else if len(parts) >= 7 && parts[6] != "" {
			res.Artist = parts[6]
		}
		if len(parts) >= 4 && parts[3] != "" {
			res.Album = parts[3]
		}
		if len(parts) >= 5 && parts[4] != "" {
			sec, _ := strconv.Atoi(parts[4])
			res.Duration = sec
		}
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
