package validator

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPassword validates password strength
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z':
			hasLetter = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	return hasLetter && hasNumber
}

// IsValidURL validates URL format
func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// Check if URL has scheme, if not add http://
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check if host is valid
	if u.Host == "" || u.Scheme == "" {
		return false
	}

	// Check for common invalid patterns
	if strings.Contains(u.Host, " ") || strings.Contains(u.Host, "..") || strings.Contains(u.Host, "'") {
		return false
	}

	return true
}

// IsValidShortCode validates short code format
func IsValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	// Only alphanumeric characters, hyphens, and underscores
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}
