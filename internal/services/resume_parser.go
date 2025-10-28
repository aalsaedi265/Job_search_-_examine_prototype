package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/yourusername/jobapply/internal/models"
)

// ParseResume extracts work history from a PDF resume file
func ParseResume(filePath string) ([]models.WorkHistory, error) {
	// Open PDF file
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	// Extract all text from PDF
	var textContent strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		textContent.WriteString(text)
		textContent.WriteString("\n")
	}

	resumeText := textContent.String()

	// Parse work history from text
	workHistory := extractWorkHistory(resumeText)

	return workHistory, nil
}

// extractWorkHistory parses resume text to extract work experience entries
func extractWorkHistory(text string) []models.WorkHistory {
	var workHistory []models.WorkHistory

	// Common section headers for work experience
	experienceHeaders := []string{
		"WORK EXPERIENCE", "PROFESSIONAL EXPERIENCE", "EXPERIENCE",
		"EMPLOYMENT HISTORY", "WORK HISTORY", "CAREER HISTORY",
	}

	// Find the work experience section
	text = strings.ToUpper(text)
	var experienceSection string
	for _, header := range experienceHeaders {
		idx := strings.Index(text, header)
		if idx != -1 {
			// Extract text from this section until next major section
			endHeaders := []string{"EDUCATION", "SKILLS", "CERTIFICATIONS", "PROJECTS", "AWARDS"}
			endIdx := len(text)
			for _, endHeader := range endHeaders {
				if endHeaderIdx := strings.Index(text[idx:], endHeader); endHeaderIdx != -1 && endHeaderIdx < endIdx {
					endIdx = idx + endHeaderIdx
					break
				}
			}
			experienceSection = text[idx:endIdx]
			break
		}
	}

	if experienceSection == "" {
		// If no clear section found, try to extract from entire resume
		experienceSection = text
	}

	// Parse work entries using date patterns and heuristics
	workHistory = parseWorkEntries(experienceSection)

	return workHistory
}

// parseWorkEntries extracts individual work experience entries from text
func parseWorkEntries(text string) []models.WorkHistory {
	var workHistory []models.WorkHistory

	// First, try to find pipe-separated single-line format
	// Format: "Company | Title | Extra Info | MM/YYYY - MM/YYYY" or "Company | Title | MM/YYYY - present"
	lines := strings.Split(text, "\n")

	// Pattern for dates at end of line
	dateRangePattern := regexp.MustCompile(`(\d{1,2}/\d{4}|\w+\s+\d{4}|\d{4})\s*[-–—]\s*(\d{1,2}/\d{4}|\w+\s+\d{4}|\d{4}|PRESENT|CURRENT)\s*$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check if line contains pipes and a date range
		if strings.Count(line, "|") >= 2 && dateRangePattern.MatchString(line) {
			// Parse single-line format: "Company | Title | Extra | Dates"
			entry := parseSingleLineEntry(line)
			if entry.Company != "" || entry.Title != "" {
				// Get description from following lines until next entry
				description := extractDescriptionFromLines(lines, i+1)
				entry.Description = cleanDescription(description)
				workHistory = append(workHistory, entry)
				continue
			}
		}
	}

	// If single-line parsing didn't find much, fall back to multi-line parsing
	if len(workHistory) == 0 {
		workHistory = parseMultiLineEntries(text)
	}

	return workHistory
}

// parseSingleLineEntry parses format: "Company | Title | Extra Info | MM/YYYY - MM/YYYY"
func parseSingleLineEntry(line string) models.WorkHistory {
	var entry models.WorkHistory

	// Extract date range from end of line
	dateRangePattern := regexp.MustCompile(`(\d{1,2}/\d{4}|\w+\s+\d{4}|\d{4})\s*[-–—]\s*(\d{1,2}/\d{4}|\w+\s+\d{4}|\d{4}|PRESENT|CURRENT)\s*$`)
	dateMatch := dateRangePattern.FindStringSubmatch(line)

	if dateMatch == nil {
		return entry
	}

	startDate := normalizeDate(dateMatch[1])
	endDate := normalizeDate(dateMatch[2])

	// Remove date range from line
	line = dateRangePattern.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)

	// Split by pipe separator
	parts := strings.Split(line, "|")

	// Clean up parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Parse based on number of parts
	if len(parts) >= 2 {
		entry.Company = parts[0]
		entry.Title = parts[1]

		// If there are extra parts (like "Managing 4 direct reports"), append to title
		if len(parts) > 2 {
			for i := 2; i < len(parts); i++ {
				if parts[i] != "" {
					entry.Title += " | " + parts[i]
				}
			}
		}
	} else if len(parts) == 1 {
		entry.Title = parts[0]
	}

	entry.StartDate = startDate
	entry.EndDate = endDate

	return entry
}

// extractDescriptionFromLines gets description text from following lines
func extractDescriptionFromLines(lines []string, startIdx int) string {
	var descLines []string

	// Pattern to detect next entry (date range at end of line with pipes)
	nextEntryPattern := regexp.MustCompile(`\|\s*(\d{1,2}/\d{4}|\w+\s+\d{4}|\d{4})\s*[-–—]`)

	for i := startIdx; i < len(lines) && i < startIdx+20; i++ {
		line := strings.TrimSpace(lines[i])

		// Stop if we hit next entry
		if strings.Count(line, "|") >= 2 && nextEntryPattern.MatchString(line) {
			break
		}

		// Stop at section headers
		if len(line) > 0 && line == strings.ToUpper(line) && !strings.HasPrefix(line, "•") && len(line) < 50 {
			// This might be a section header within the description, include it
			descLines = append(descLines, line)
		} else if line != "" {
			descLines = append(descLines, line)
		}
	}

	return strings.Join(descLines, "\n")
}

// parseMultiLineEntries is fallback for multi-line format
func parseMultiLineEntries(text string) []models.WorkHistory {
	var workHistory []models.WorkHistory

	// Pattern to match date ranges (e.g., "Jan 2020 - Dec 2022", "2020-2022", "01/2020 - 12/2022")
	dateRangePattern := regexp.MustCompile(`(?i)(\w+\s+\d{4}|\d{1,2}/\d{4}|\d{4})\s*[-–—]\s*(\w+\s+\d{4}|\d{1,2}/\d{4}|\d{4}|PRESENT|CURRENT)`)

	// Find all date ranges
	dateMatches := dateRangePattern.FindAllStringSubmatchIndex(text, -1)

	// For each date range, extract the surrounding context
	for i, match := range dateMatches {
		startDate := text[match[2]:match[3]]
		endDate := text[match[4]:match[5]]

		// Normalize dates to YYYY-MM-DD format
		startDateNorm := normalizeDate(startDate)
		endDateNorm := normalizeDate(endDate)

		// Extract job title and company from text before the date
		// Use larger context window to capture full company/title info
		contextStart := match[0] - 500
		if contextStart < 0 {
			contextStart = 0
		}

		// Get next entry start position for description boundary
		nextEntryStart := len(text)
		if i < len(dateMatches)-1 {
			nextEntryStart = dateMatches[i+1][0] - 200 // Leave some buffer
		}

		contextText := text[contextStart:match[0]]
		descriptionText := text[match[1]:nextEntryStart]

		// Extract company and title from context
		lines := strings.Split(strings.TrimSpace(contextText), "\n")
		var company, title string

		// Filter lines to find candidates (remove bullet points, section headers, etc.)
		candidateLines := make([]string, 0)
		skipWords := []string{"EXPERIENCE", "WORK HISTORY", "EMPLOYMENT", "PROFESSIONAL", "REPORTS", "PAGE"}

		for j := len(lines) - 1; j >= 0 && len(candidateLines) < 10; j-- {
			line := strings.TrimSpace(lines[j])

			// Skip empty lines, bullet points, and section headers
			if line == "" || strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") ||
			   strings.HasPrefix(line, "●") || len(line) < 3 {
				continue
			}

			// Skip common section headers and junk
			shouldSkip := false
			for _, skip := range skipWords {
				if strings.Contains(line, skip) && len(line) < 30 {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue
			}

			candidateLines = append([]string{line}, candidateLines...)
		}

		// Try to find pipe-separated format first: "Company | Title"
		foundPipeFormat := false
		for _, line := range candidateLines {
			if strings.Contains(line, " | ") {
				parts := strings.Split(line, " | ")
				if len(parts) >= 2 {
					company = strings.TrimSpace(parts[0])
					title = strings.TrimSpace(strings.Join(parts[1:], " | "))
					foundPipeFormat = true
					break
				}
			}
		}

		// If no pipe format, use heuristic: last 1-2 significant lines before date
		if !foundPipeFormat && len(candidateLines) > 0 {
			// If we have 2+ lines, assume first is company, second is title
			if len(candidateLines) >= 2 {
				// Check if first line looks like a company (often shorter, proper case)
				line1 := candidateLines[len(candidateLines)-2]
				line2 := candidateLines[len(candidateLines)-1]

				// Heuristic: company name is usually shorter than title
				if len(line1) < len(line2) && len(line1) < 50 {
					company = line1
					title = line2
				} else {
					title = line1
					company = line2
				}
			} else {
				// Only one line - treat as title
				title = candidateLines[0]
			}
		}

		// Extract description (bullet points after date)
		description := strings.TrimSpace(descriptionText)

		// Clean up description: remove excessive whitespace
		description = cleanDescription(description)

		// Only add if we have meaningful data
		if (title != "" || company != "") && startDateNorm != "" {
			workHistory = append(workHistory, models.WorkHistory{
				Company:     company,
				Title:       title,
				StartDate:   startDateNorm,
				EndDate:     endDateNorm,
				Description: description,
			})
		}
	}

	return workHistory
}

// normalizeDate converts various date formats to YYYY-MM-DD
func normalizeDate(dateStr string) string {
	originalDateStr := dateStr
	dateStr = strings.TrimSpace(strings.ToUpper(dateStr))

	// Handle "PRESENT" or "CURRENT"
	if strings.Contains(dateStr, "PRESENT") || strings.Contains(dateStr, "CURRENT") {
		fmt.Printf("[DEBUG] Date conversion: '%s' -> '' (present)\n", originalDateStr)
		return "" // Empty string indicates current job
	}

	// Month name patterns
	monthMap := map[string]string{
		"JAN": "01", "JANUARY": "01",
		"FEB": "02", "FEBRUARY": "02",
		"MAR": "03", "MARCH": "03",
		"APR": "04", "APRIL": "04",
		"MAY": "05",
		"JUN": "06", "JUNE": "06",
		"JUL": "07", "JULY": "07",
		"AUG": "08", "AUGUST": "08",
		"SEP": "09", "SEPTEMBER": "09",
		"OCT": "10", "OCTOBER": "10",
		"NOV": "11", "NOVEMBER": "11",
		"DEC": "12", "DECEMBER": "12",
	}

	// Try to parse "Mon YYYY" or "Month YYYY"
	monthYearPattern := regexp.MustCompile(`(\w+)\s+(\d{4})`)
	if match := monthYearPattern.FindStringSubmatch(dateStr); match != nil {
		month := strings.ToUpper(match[1])
		year := match[2]
		if monthNum, ok := monthMap[month]; ok {
			result := fmt.Sprintf("%s-%s-01", year, monthNum)
			fmt.Printf("[DEBUG] Date conversion: '%s' -> '%s' (month name)\n", originalDateStr, result)
			return result
		}
	}

	// Try to parse "MM/YYYY" or "M/YYYY"
	mmYYYYPattern := regexp.MustCompile(`^(\d{1,2})/(\d{4})$`)
	if match := mmYYYYPattern.FindStringSubmatch(dateStr); match != nil {
		month := match[1]
		year := match[2]
		if len(month) == 1 {
			month = "0" + month
		}
		result := fmt.Sprintf("%s-%s-01", year, month)
		fmt.Printf("[DEBUG] Date conversion: '%s' -> '%s' (MM/YYYY)\n", originalDateStr, result)
		return result
	}

	// Try to parse just "YYYY"
	yyyyPattern := regexp.MustCompile(`^(\d{4})$`)
	if match := yyyyPattern.FindStringSubmatch(dateStr); match != nil {
		result := fmt.Sprintf("%s-01-01", match[1])
		fmt.Printf("[DEBUG] Date conversion: '%s' -> '%s' (YYYY)\n", originalDateStr, result)
		return result
	}

	// If we can't parse, return as-is (backend will handle validation)
	fmt.Printf("[DEBUG] Date conversion: '%s' -> '%s' (unchanged)\n", originalDateStr, dateStr)
	return dateStr
}

// ExtractResumeText extracts plain text from PDF for debugging
func ExtractResumeText(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var textContent strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		textContent.WriteString(fmt.Sprintf("=== Page %d ===\n", pageNum))
		textContent.WriteString(text)
		textContent.WriteString("\n\n")
	}

	return textContent.String(), nil
}

// cleanDescription removes excessive whitespace and formats text properly
func cleanDescription(text string) string {
	// First, normalize all whitespace - this handles the case where PDF extraction
	// puts spaces/newlines between every word
	spacePattern := regexp.MustCompile(`\s+`)
	text = spacePattern.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Now intelligently split into proper lines
	// Look for bullet points and section headers
	text = strings.ReplaceAll(text, "●", "\n• ")
	text = strings.ReplaceAll(text, "•", "\n• ")

	// Find section headers (all caps, 10+ chars) and add newlines
	// Split into words and identify sequences of uppercase words
	lines := strings.Split(text, "\n")
	cleanedLines := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line looks like a section header (mostly uppercase, 10+ chars)
		if len(line) >= 10 {
			upperCount := 0
			for _, ch := range line {
				if ch >= 'A' && ch <= 'Z' || ch == ' ' || ch == '&' {
					upperCount++
				}
			}
			// If more than 80% uppercase/spaces/&, it's likely a header
			if float64(upperCount)/float64(len(line)) > 0.8 {
				// Add extra newline before header for separation
				if len(cleanedLines) > 0 {
					cleanedLines = append(cleanedLines, "")
				}
			}
		}

		cleanedLines = append(cleanedLines, line)
	}

	text = strings.Join(cleanedLines, "\n")

	// Clean up excessive newlines (max 2 consecutive)
	newlinePattern := regexp.MustCompile(`\n{3,}`)
	text = newlinePattern.ReplaceAllString(text, "\n\n")

	// Trim and return
	return strings.TrimSpace(text)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
