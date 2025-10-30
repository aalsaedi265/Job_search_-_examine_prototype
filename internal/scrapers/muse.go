package scrapers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Job struct {
	Title    string
	Company  string
	Location string
	URL      string
}

type MuseScraper struct {
	client *http.Client
}

func NewMuseScraper() *MuseScraper {
	return &MuseScraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Muse API response structures
type museResponse struct {
	Results []museJob `json:"results"`
	Page    int       `json:"page"`
}

type museJob struct {
	Name     string      `json:"name"`     // Job title
	Company  museCompany `json:"company"`  // Company info
	Locations []museLocation `json:"locations"` // Job locations
	Refs     museRefs    `json:"refs"`     // URLs
}

type museCompany struct {
	Name string `json:"name"`
}

type museLocation struct {
	Name string `json:"name"` // e.g., "New York, NY"
}

type museRefs struct {
	LandingPage string `json:"landing_page"` // Application URL
}

func (s *MuseScraper) Scrape(keywords, location string) ([]Job, error) {
	// Build The Muse API URL
	baseURL := "https://www.themuse.com/api/public/jobs"
	params := url.Values{}

	// Muse API only supports category (broad) and location filters
	// Categories: "Software Engineer", "Data Science", etc.
	if keywords != "" {
		params.Add("category", keywords)
	}
	if location != "" {
		params.Add("location", location)
	}
	params.Add("page", "0")
	params.Add("descending", "true")

	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	fmt.Printf("[DEBUG] Muse API URL: %s\n", apiURL)

	// Make HTTP request
	resp, err := s.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var museResp museResponse
	if err := json.NewDecoder(resp.Body).Decode(&museResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("[DEBUG] Muse API returned %d jobs\n", len(museResp.Results))

	// Convert to our Job format
	jobs := make([]Job, 0, len(museResp.Results))
	for _, mj := range museResp.Results {
		// Skip jobs without required fields
		if mj.Name == "" || mj.Company.Name == "" || mj.Refs.LandingPage == "" {
			continue
		}

		// Get first location if available
		locationStr := ""
		if len(mj.Locations) > 0 {
			locationStr = mj.Locations[0].Name
		}

		jobs = append(jobs, Job{
			Title:    mj.Name,
			Company:  mj.Company.Name,
			Location: locationStr,
			URL:      mj.Refs.LandingPage,
		})
	}

	fmt.Printf("[DEBUG] Converted %d valid jobs\n", len(jobs))
	return jobs, nil
}
