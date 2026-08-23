package ytdlp

import (
	"fmt"
	"testing"
	"time"
)

func TestValidateCookieBytes_Valid(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour).Unix()
	content := fmt.Sprintf(`# Netscape HTTP Cookie File
.youtube.com	TRUE	/	TRUE	%d	LOGIN_INFO	AFmmF2swRQIh...
.youtube.com	TRUE	/	TRUE	%d	SAPISID	valid_sapisid_token
.youtube.com	TRUE	/	TRUE	%d	__Secure-1PSID	valid_psid_token
`, future, future, future)

	res := ValidateCookieBytes([]byte(content))
	if !res.IsValid {
		t.Fatalf("expected valid cookies, got invalid: %s", res.ErrorReason)
	}
	if res.Status != CookieStatusValid {
		t.Fatalf("expected status 'valid', got '%s'", res.Status)
	}
	if !res.HasCookies {
		t.Fatalf("expected HasCookies to be true")
	}
}

func TestValidateCookieBytes_Expired(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Unix()
	content := fmt.Sprintf(`# Netscape HTTP Cookie File
.youtube.com	TRUE	/	TRUE	%d	LOGIN_INFO	expired_info
.youtube.com	TRUE	/	TRUE	%d	SID	expired_sid
`, past, past)

	res := ValidateCookieBytes([]byte(content))
	if res.IsValid {
		t.Fatalf("expected invalid/expired cookies, got valid")
	}
	if res.Status != CookieStatusExpired {
		t.Fatalf("expected status 'expired', got '%s'", res.Status)
	}
}

func TestValidateCookieBytes_MissingAuth(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour).Unix()
	content := fmt.Sprintf(`# Netscape HTTP Cookie File
.youtube.com	TRUE	/	TRUE	%d	PREF	f1=50000000
.youtube.com	TRUE	/	TRUE	%d	GPS	1
`, future, future)

	res := ValidateCookieBytes([]byte(content))
	if res.IsValid {
		t.Fatalf("expected invalid cookies without auth tokens, got valid")
	}
	if res.Status != CookieStatusInvalid {
		t.Fatalf("expected status 'invalid', got '%s'", res.Status)
	}
}

func TestIsYTDLPAuthError(t *testing.T) {
	tests := []struct {
		stderr   string
		expected bool
	}{
		{"ERROR: Sign in to confirm you're not a bot. Use --cookies-from-browser", true},
		{"ERROR: [youtube] Private video. Sign in if you've been granted access", true},
		{"ERROR: [youtube] cookies are expired", true},
		{"ERROR: [youtube] 429 Too Many Requests", false},
		{"[download] 100% of 3.20MiB in 00:01", false},
	}

	for _, tt := range tests {
		got, _ := IsYTDLPAuthError(tt.stderr)
		if got != tt.expected {
			t.Errorf("IsYTDLPAuthError(%q) = %v, expected %v", tt.stderr, got, tt.expected)
		}
	}
}
