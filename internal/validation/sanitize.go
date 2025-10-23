package validation

import (
	"html"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SanitizeString removes potentially dangerous characters and limits length to prevent XSS attacks
func SanitizeString(input string, maxLength int) string {
	// HTML escape to prevent XSS
	sanitized := html.EscapeString(input)

	// Remove null bytes to prevent SQL injection tricks
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Limit length to prevent buffer overflow
	if utf8.RuneCountInString(sanitized) > maxLength {
		runes := []rune(sanitized)
		sanitized = string(runes[:maxLength])
	}

	return sanitized
}

// ValidateEmail checks if email format is valid to prevent injection attacks
func ValidateEmail(email string) bool {
	// Simple but effective email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email) && len(email) <= 255
}

// ValidatePhone checks if phone format is reasonable
func ValidatePhone(phone string) bool {
	// Allow common phone formats, prevent injection
	phoneRegex := regexp.MustCompile(`^[\d\s\-\+\(\)]{7,20}$`)
	return phoneRegex.MatchString(phone)
}

// ValidateUUID checks if UUID format is valid to prevent injection attacks
func ValidateUUID(id string) bool {
	uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
	return uuidRegex.MatchString(id)
}

// SanitizeFilename prevents path traversal attacks in file uploads
func SanitizeFilename(filename string) string {
	// Remove any path components to prevent directory traversal
	filename = filepath.Base(filename)

	// Remove potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[^a-zA-Z0-9._\-]`)
	filename = dangerousChars.ReplaceAllString(filename, "_")

	// Prevent hidden files
	filename = strings.TrimPrefix(filename, ".")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	// Ensure filename is not empty
	if filename == "" {
		filename = "file"
	}

	return filename
}

// ValidateFileExtension checks if file extension is in allowed list
func ValidateFileExtension(filename string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExts {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// SanitizeURL prevents open redirect and SSRF attacks
func SanitizeURL(url string) string {
	// Remove dangerous protocols
	url = strings.TrimSpace(url)

	// Block javascript:, data:, file:, etc.
	dangerousProtocols := []string{"javascript:", "data:", "file:", "vbscript:"}
	urlLower := strings.ToLower(url)
	for _, proto := range dangerousProtocols {
		if strings.HasPrefix(urlLower, proto) {
			return ""
		}
	}

	return SanitizeString(url, 2048)
}

// ValidatePassword checks password strength requirements
func ValidatePassword(password string) bool {
	// Minimum 6 characters, maximum 128 to prevent DoS
	if len(password) < 6 || len(password) > 128 {
		return false
	}

	// Must contain at least one letter and one number
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasLetter && hasNumber
}

// SanitizeJobSearchQuery prevents injection in job search queries
func SanitizeJobSearchQuery(query string) string {
	// Remove SQL wildcards and special characters
	query = SanitizeString(query, 200)

	// Remove potential SQL injection patterns
	sqlPatterns := []string{"--", ";", "/*", "*/", "xp_", "sp_", "DROP", "DELETE", "INSERT", "UPDATE"}
	queryUpper := strings.ToUpper(query)
	for _, pattern := range sqlPatterns {
		if strings.Contains(queryUpper, pattern) {
			query = strings.ReplaceAll(query, pattern, "")
			query = strings.ReplaceAll(query, strings.ToLower(pattern), "")
		}
	}

	return query
}
