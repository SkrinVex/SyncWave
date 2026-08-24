package ytdlp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CookieStatus string

const (
	CookieStatusValid        CookieStatus = "valid"
	CookieStatusExpiringSoon CookieStatus = "expiring_soon"
	CookieStatusExpired      CookieStatus = "expired"
	CookieStatusMissing      CookieStatus = "missing"
	CookieStatusInvalid      CookieStatus = "invalid"
)

type CookieValidationResult struct {
	Status        CookieStatus `json:"status"`
	IsValid       bool         `json:"is_valid"`
	HasCookies    bool         `json:"has_cookies"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	ExpiresInDays int          `json:"expires_in_days"`
	ErrorReason   string       `json:"error_reason,omitempty"`
}

// Core YouTube authentication cookie names
var coreAuthCookies = map[string]bool{
	"LOGIN_INFO":     true,
	"SAPISID":        true,
	"__Secure-1PSID": true,
	"__Secure-3PSID": true,
	"SID":            true,
	"SSID":           true,
	"HSID":           true,
}

// ValidateCookieFile parses and validates a Netscape cookies.txt file on disk
func ValidateCookieFile(filePath string) *CookieValidationResult {
	if filePath == "" {
		return &CookieValidationResult{
			Status:      CookieStatusMissing,
			IsValid:     false,
			HasCookies:  false,
			ErrorReason: "Файл cookies.txt не загружен",
		}
	}

	info, err := os.Stat(filePath)
	if err != nil || info.Size() == 0 {
		return &CookieValidationResult{
			Status:      CookieStatusMissing,
			IsValid:     false,
			HasCookies:  false,
			ErrorReason: "Файл cookies.txt пуст или отсутствует",
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return &CookieValidationResult{
			Status:      CookieStatusInvalid,
			IsValid:     false,
			HasCookies:  false,
			ErrorReason: fmt.Sprintf("Ошибка чтения cookies.txt: %v", err),
		}
	}

	return ValidateCookieBytes(content)
}

// ValidateCookieBytes parses Netscape format bytes and checks expiration of auth tokens
func ValidateCookieBytes(content []byte) *CookieValidationResult {
	if len(content) == 0 {
		return &CookieValidationResult{
			Status:      CookieStatusMissing,
			IsValid:     false,
			HasCookies:  false,
			ErrorReason: "Файл cookies пуст",
		}
	}

	content = NormalizeCookiesToNetscape(content)
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var foundAuthCookies []string
	var earliestExpiry *time.Time
	var expiredCookieNames []string

	now := time.Now().UTC()
	lineCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			// Tab-separated fallback
			parts = strings.Split(line, "\t")
			if len(parts) < 6 {
				continue
			}
		}

		domain := parts[0]
		// Check only youtube / google domain cookies
		if !strings.Contains(domain, "youtube.com") && !strings.Contains(domain, "google.com") && !strings.Contains(domain, ".google.") {
			continue
		}

		name := ""
		expiryIdx := 4
		nameIdx := 5

		if len(parts) >= 7 {
			name = parts[nameIdx]
		} else if len(parts) >= 6 {
			name = parts[nameIdx]
		}

		if coreAuthCookies[name] {
			foundAuthCookies = append(foundAuthCookies, name)

			if len(parts) > expiryIdx {
				expSec, err := strconv.ParseInt(parts[expiryIdx], 10, 64)
				if err == nil && expSec > 0 {
					expTime := time.Unix(expSec, 0).UTC()
					if expTime.Before(now) {
						expiredCookieNames = append(expiredCookieNames, fmt.Sprintf("%s (истёк %s)", name, expTime.Format("02.01.2006")))
					} else {
						if earliestExpiry == nil || expTime.Before(*earliestExpiry) {
							earliestExpiry = &expTime
						}
					}
				}
			}
		}
	}

	if len(foundAuthCookies) == 0 {
		return &CookieValidationResult{
			Status:      CookieStatusInvalid,
			IsValid:     false,
			HasCookies:  true,
			ErrorReason: "Файл cookies.txt не содержит авторизационных данных YouTube (LOGIN_INFO / SID / SAPISID). Экспортируйте куки из аккаунта YouTube Music.",
		}
	}

	if len(expiredCookieNames) > 0 && earliestExpiry == nil {
		return &CookieValidationResult{
			Status:      CookieStatusExpired,
			IsValid:     false,
			HasCookies:  true,
			ErrorReason: fmt.Sprintf("Срок действия сессии YouTube истёк: %s. Пожалуйста, обновите cookies.txt.", strings.Join(expiredCookieNames, ", ")),
		}
	}

	daysRemaining := 0
	if earliestExpiry != nil {
		daysRemaining = int(time.Until(*earliestExpiry).Hours() / 24)
	}

	if daysRemaining < 7 && daysRemaining >= 0 {
		return &CookieValidationResult{
			Status:        CookieStatusExpiringSoon,
			IsValid:       true,
			HasCookies:    true,
			ExpiresAt:     earliestExpiry,
			ExpiresInDays: daysRemaining,
			ErrorReason:   fmt.Sprintf("Срок действия cookies истекает через %d дн. Рекомендуется обновить cookies.txt.", daysRemaining),
		}
	}

	return &CookieValidationResult{
		Status:        CookieStatusValid,
		IsValid:       true,
		HasCookies:    true,
		ExpiresAt:     earliestExpiry,
		ExpiresInDays: daysRemaining,
	}
}

// NormalizeCookiesToNetscape ensures the cookie content is in standard Netscape format for yt-dlp
func NormalizeCookiesToNetscape(content []byte) []byte {
	str := strings.TrimSpace(string(content))
	if str == "" {
		return content
	}

	// 1. If already in Netscape format
	if strings.Contains(str, "# Netscape HTTP Cookie File") || strings.Contains(str, "\tTRUE\t") || strings.Contains(str, "\tFALSE\t") {
		return content
	}

	// 2. Check if JSON array format
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var jsonCookies []struct {
			Domain     string      `json:"domain"`
			Expiration interface{} `json:"expirationDate"`
			HostOnly   bool        `json:"hostOnly"`
			HttpOnly   bool        `json:"httpOnly"`
			Name       string      `json:"name"`
			Path       string      `json:"path"`
			Secure     bool        `json:"secure"`
			Value      string      `json:"value"`
		}
		if err := json.Unmarshal(content, &jsonCookies); err == nil && len(jsonCookies) > 0 {
			var buf bytes.Buffer
			buf.WriteString("# Netscape HTTP Cookie File\n# http://curl.haxx.se/rfc/cookie_spec.html\n# This is a generated file!  Do not edit.\n\n")
			for _, c := range jsonCookies {
				exp := int64(2147483647)
				if c.Expiration != nil {
					switch v := c.Expiration.(type) {
					case float64:
						exp = int64(v)
					case int64:
						exp = v
					case string:
						if parsed, pErr := strconv.ParseInt(v, 10, 64); pErr == nil {
							exp = parsed
						}
					}
				}
				dom := c.Domain
				if dom == "" {
					dom = ".youtube.com"
				}
				path := c.Path
				if path == "" {
					path = "/"
				}
				subdomains := "TRUE"
				if strings.HasPrefix(dom, ".") {
					subdomains = "TRUE"
				} else {
					subdomains = "FALSE"
				}
				secure := "FALSE"
				if c.Secure {
					secure = "TRUE"
				}
				buf.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s\n", dom, subdomains, path, secure, exp, c.Name, c.Value))
			}
			return buf.Bytes()
		}
	}

	// 3. Raw header format: "LOGIN_INFO=xxx; SAPISID=yyy; SID=zzz; ..."
	if strings.Contains(str, "=") {
		pairs := strings.Split(str, ";")
		var buf bytes.Buffer
		buf.WriteString("# Netscape HTTP Cookie File\n# http://curl.haxx.se/rfc/cookie_spec.html\n# This is a generated file!  Do not edit.\n\n")
		expiry := time.Now().Add(365 * 24 * time.Hour).Unix()
		count := 0

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			eqIdx := strings.Index(pair, "=")
			if eqIdx == -1 {
				continue
			}
			name := strings.TrimSpace(pair[:eqIdx])
			value := strings.TrimSpace(pair[eqIdx+1:])

			if name == "" {
				continue
			}

			// Add for youtube.com
			buf.WriteString(fmt.Sprintf(".youtube.com\tTRUE\t/\tTRUE\t%d\t%s\t%s\n", expiry, name, value))
			// Also add for google.com
			buf.WriteString(fmt.Sprintf(".google.com\tTRUE\t/\tTRUE\t%d\t%s\t%s\n", expiry, name, value))
			count++
		}

		if count > 0 {
			return buf.Bytes()
		}
	}

	return content
}

// IsYTDLPAuthError inspects yt-dlp error output for authentication and expired cookie indicators
func IsYTDLPAuthError(stderr string) (bool, string) {
	lower := strings.ToLower(stderr)

	if strings.Contains(lower, "sign in to confirm you're not a bot") ||
		strings.Contains(lower, "sign in to confirm you’re not a bot") ||
		strings.Contains(lower, "confirm you're not a bot") {
		return true, "YouTube требует авторизацию (Sign in to confirm you're not a bot). Пожалуйста, обновите cookies.txt."
	}

	if strings.Contains(lower, "private video") ||
		strings.Contains(lower, "sign in if you've been granted access") ||
		strings.Contains(lower, "sign in if you’ve been granted access") {
		return true, "Приватное видео или закрытый плейлист. Требуется актуальный файл cookies.txt."
	}

	if strings.Contains(lower, "sign in to confirm your age") ||
		strings.Contains(lower, "age-restricted") {
		return true, "Видео с возрастным ограничением 18+. Требуется авторизация через cookies.txt."
	}

	if strings.Contains(lower, "the following content is not available on this app") {
		return true, "YouTube заблокировал клиент приложения. Требуется обновить cookies.txt или настроить прокси."
	}

	if strings.Contains(lower, "cookies are expired") ||
		strings.Contains(lower, "cookie has expired") {
		return true, "Срок действия cookies YouTube истёк. Пожалуйста, обновите cookies.txt."
	}

	if strings.Contains(lower, "http error 401") ||
		strings.Contains(lower, "http error 403: forbidden") {
		return true, "Ошибка доступа к YouTube (HTTP 401/403). Требуется обновить cookies.txt или сменить прокси."
	}

	return false, ""
}
