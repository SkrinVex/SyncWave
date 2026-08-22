package ytdlp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	ID       string `json:"id"`
	Title    string `json:"title"`
	Uploader string `json:"uploader"`
	Artist   string `json:"artist"`
	Track    string `json:"track"`
	Duration int    `json:"duration"`
	URL      string `json:"url"`
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
	if format == "" {
		format = "opus"
	}
	c.audioFormat = format
}

func (c *Client) GetYTDLPVersion() (string, error) {
	cmd := exec.Command(c.ytdlpPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) GetFFmpegVersion() (string, error) {
	cmd := exec.Command(c.ffmpegPath, "-version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
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

	if c.ffmpegPath != "" {
		args = append(args, "--ffmpeg-location", c.ffmpegPath)
	}

	if c.HasCookies() {
		args = append(args, "--cookies", c.cookiesPath)
	}

	if c.proxyURL != "" {
		args = append(args, "--proxy", c.proxyURL)
	}

	// Browser/client evasion args
	args = append(args,
		"--extractor-args", "youtube:player_client=android,web",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	)

	return args
}

// NormalizePlaylistURL converts "LM", playlist IDs, or full URLs into standard yt-dlp target URLs
func NormalizePlaylistURL(rawInput string) string {
	raw := strings.TrimSpace(rawInput)
	if raw == "LM" || raw == "liked" || raw == "likes" {
		return "https://music.youtube.com/playlist?list=LM"
	}
	if strings.HasPrefix(raw, "PL") || strings.HasPrefix(raw, "RD") || strings.HasPrefix(raw, "OL") || strings.HasPrefix(raw, "VL") {
		return "https://music.youtube.com/playlist?list=" + raw
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return "https://music.youtube.com/playlist?list=" + raw
	}
	return raw
}

// ExtractPlaylistDelta fetches fast JSON metadata (--flat-playlist) without downloading media streams
func (c *Client) ExtractPlaylistDelta(ctx context.Context, playlistInput string) (*FlatPlaylistOutput, error) {
	url := NormalizePlaylistURL(playlistInput)

	args := append(c.buildBaseArgs(),
		"--flat-playlist",
		"-J",
		"--ignore-errors",
		"--no-warnings",
		url,
	)

	cmd := exec.CommandContext(ctx, c.ytdlpPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract playlist: %v, stderr: %s", err, stderr.String())
	}

	var output FlatPlaylistOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		// yt-dlp might return single video or alternate JSON layout
		var singleEntry PlaylistEntry
		if err2 := json.Unmarshal(stdout.Bytes(), &singleEntry); err2 == nil && singleEntry.ID != "" {
			return &FlatPlaylistOutput{
				ID:      singleEntry.ID,
				Title:   singleEntry.Title,
				Entries: []PlaylistEntry{singleEntry},
			}, nil
		}
		return nil, fmt.Errorf("failed to parse playlist json: %w", err)
	}

	return &output, nil
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
	progressRegex = regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+~?([^\s]+)\s+at\s+([^\s]+)\s+ETA\s+([^\s]+)`)
)

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

	args := append(c.buildBaseArgs(),
		"-x",
		"--audio-format", format,
		"--audio-quality", "0",
		"--embed-thumbnail",
		"--add-metadata",
		"--write-thumbnail",
		"--convert-thumbnails", "jpg",
		"--no-playlist",
		"-o", outTemplate,
		"--print-json",
		"--newline",
		targetURL,
	)

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

	var jsonOutput bytes.Buffer
	scanner := bufio.NewScanner(stdoutPipe)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line is json metadata output from --print-json
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			jsonOutput.WriteString(line)
			continue
		}

		// Parse download progress
		if onProgress != nil {
			if matches := progressRegex.FindStringSubmatch(line); len(matches) == 5 {
				percent, _ := strconv.ParseFloat(matches[1], 64)
				speed := matches[3]
				eta := matches[4]
				onProgress(percent, speed, eta, "downloading")
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v, stderr: %s", err, stderr.String())
	}

	// Locate downloaded audio file
	expectedAudioPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.%s", youtubeID, format))
	fileInfo, err := os.Stat(expectedAudioPath)
	if err != nil {
		// Try finding any extension if opus converted or defaulted
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
		// Move or copy to covers directory
		_ = os.Rename(tempThumbnailPath, expectedCoverPath)
	} else {
		// Check for .webp in music dir
		tempWebpPath := filepath.Join(c.musicDir, fmt.Sprintf("%s.webp", youtubeID))
		if _, err := os.Stat(tempWebpPath); err == nil {
			// Convert webp to jpg via ffmpeg
			c.convertWebpToJpg(ctx, tempWebpPath, expectedCoverPath)
			_ = os.Remove(tempWebpPath)
		}
	}

	// Clean up any remaining temporary image files in music directory
	for _, ext := range []string{".jpg", ".webp", ".png"} {
		_ = os.Remove(filepath.Join(c.musicDir, fmt.Sprintf("%s%s", youtubeID, ext)))
	}

	// Parse JSON output if present for artist/title/album/duration
	title := youtubeID
	artist := "Unknown Artist"
	album := "Unknown Album"
	duration := 0

	if jsonOutput.Len() > 0 {
		var meta struct {
			Title    string      `json:"title"`
			Artist   string      `json:"artist"`
			Creator  string      `json:"creator"`
			Uploader string      `json:"uploader"`
			Channel  string      `json:"channel"`
			Album    string      `json:"album"`
			Duration interface{} `json:"duration"`
			Track    string      `json:"track"`
		}
		if err := json.Unmarshal(jsonOutput.Bytes(), &meta); err == nil {
			if meta.Track != "" {
				title = meta.Track
			} else if meta.Title != "" {
				title = meta.Title
			}

			if meta.Artist != "" {
				artist = meta.Artist
			} else if meta.Creator != "" {
				artist = meta.Creator
			} else if meta.Uploader != "" {
				artist = meta.Uploader
			} else if meta.Channel != "" {
				artist = meta.Channel
			}

			if meta.Album != "" {
				album = meta.Album
			}

			switch v := meta.Duration.(type) {
			case float64:
				duration = int(v)
			case int:
				duration = v
			case string:
				duration, _ = strconv.Atoi(v)
			}
		}
	}

	// Refine title/artist if title contains "Artist - Title" pattern
	if (artist == "Unknown Artist" || strings.Contains(title, " - ")) && strings.Contains(title, " - ") {
		parts := strings.SplitN(title, " - ", 2)
		if len(parts) == 2 {
			artistCandidate := strings.TrimSpace(parts[0])
			titleCandidate := strings.TrimSpace(parts[1])
			if artistCandidate != "" && titleCandidate != "" {
				artist = artistCandidate
				title = titleCandidate
			}
		}
	}

	var coverFinalPath string
	if _, err := os.Stat(expectedCoverPath); err == nil {
		coverFinalPath = expectedCoverPath
	}

	return &DownloadResult{
		YouTubeID: youtubeID,
		Title:     title,
		Artist:    artist,
		Album:     album,
		Duration:  duration,
		FilePath:  expectedAudioPath,
		CoverPath: coverFinalPath,
		FileSize:  fileInfo.Size(),
		Format:    format,
		Bitrate:   160,
	}, nil
}

func (c *Client) convertWebpToJpg(ctx context.Context, webpPath, jpgPath string) {
	cmd := exec.CommandContext(ctx, c.ffmpegPath, "-y", "-i", webpPath, jpgPath)
	_ = cmd.Run()
}

func (c *Client) TestProxyConnection(ctx context.Context, proxyURL string) error {
	cmd := exec.CommandContext(ctx, c.ytdlpPath,
		"--proxy", proxyURL,
		"--dump-user-agent",
	)
	return cmd.Run()
}

// CopyFile helper for moving static assets
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
