package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/yourusername/jobapply/internal/scrapers"
)

type ScrapeRequest struct {
	Keywords string `json:"keywords"`
	Location string `json:"location"`
}

type ScrapeResponse struct {
	JobsScraped int  `json:"jobs_scraped"`
	FromCache   bool `json:"from_cache"`
}

// ScrapeJobs handles the POST /api/v1/scrape endpoint with caching
func (h *Handler) ScrapeJobs(w http.ResponseWriter, r *http.Request) {
	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Keywords == "" || req.Location == "" {
		h.error(w, "keywords and location are required", http.StatusBadRequest)
		return
	}

	// Generate cache key from search params
	searchHash := generateSearchHash(req.Keywords, req.Location)

	// Check cache first (jobs < 12 hours old)
	cacheQuery := `
		SELECT COUNT(*)
		FROM jobs
		WHERE search_params_hash = $1
		AND cached_at > NOW() - INTERVAL '12 hours'
	`
	var cachedCount int
	err := h.db.QueryRow(r.Context(), cacheQuery, searchHash).Scan(&cachedCount)

	if err == nil && cachedCount > 0 {
		log.Printf("Cache hit for search: %s in %s (%d jobs)", req.Keywords, req.Location, cachedCount)
		h.json(w, ScrapeResponse{
			JobsScraped: cachedCount,
			FromCache:   true,
		}, http.StatusOK)
		return
	}

	// Cache miss - fetch from Muse API
	log.Printf("Cache miss - calling Muse API for: %s in %s", req.Keywords, req.Location)

	scraper := scrapers.NewMuseScraper()
	jobs, err := scraper.Scrape(req.Keywords, req.Location)
	if err != nil {
		log.Printf("Scraping error: %v", err)
		h.error(w, "Scraping failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Scraped %d jobs from Muse API", len(jobs))

	// Insert jobs with cache metadata
	insertQuery := `
		INSERT INTO jobs (site, title, company, location, url, search_params_hash, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (url) DO UPDATE SET
			search_params_hash = EXCLUDED.search_params_hash,
			cached_at = NOW()
	`

	jobsInserted := 0
	for _, job := range jobs {
		_, err := h.db.Exec(r.Context(), insertQuery,
			"muse", job.Title, job.Company, job.Location, job.URL, searchHash)
		if err == nil {
			jobsInserted++
		}
	}

	// Clean up old cached entries (> 24 hours)
	deleteOldQuery := `
		DELETE FROM jobs
		WHERE cached_at < NOW() - INTERVAL '24 hours'
	`
	h.db.Exec(r.Context(), deleteOldQuery)

	h.json(w, ScrapeResponse{
		JobsScraped: jobsInserted,
		FromCache:   false,
	}, http.StatusOK)
}

// generateSearchHash creates a unique hash for caching
func generateSearchHash(keywords, location string) string {
	data := fmt.Sprintf("%s|%s", keywords, location)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
