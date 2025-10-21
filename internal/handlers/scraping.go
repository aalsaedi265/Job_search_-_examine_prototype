package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/yourusername/jobapply/internal/scrapers"
)

type ScrapeRequest struct {
	Keywords string `json:"keywords"`
	Location string `json:"location"`
}

type ScrapeResponse struct {
	JobsScraped int `json:"jobs_scraped"`
}

// ScrapeJobs handles the POST /api/v1/scrape endpoint
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

	// Create scraper
	scraper, err := scrapers.NewIndeedScraper()
	if err != nil {
		h.error(w, "Failed to initialize scraper", http.StatusInternalServerError)
		return
	}
	defer scraper.Close()

	// Scrape jobs
	jobs, err := scraper.Scrape(req.Keywords, req.Location)
	if err != nil {
		// Log detailed error for debugging
		log.Printf("Scraping error: %v", err)
		h.error(w, "Scraping failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log success
	log.Printf("Scraped %d jobs", len(jobs))

	// Insert jobs into database
	insertQuery := `
		INSERT INTO jobs (site, title, company, location, url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (url) DO NOTHING
	`

	jobsInserted := 0
	for _, job := range jobs {
		_, err := h.db.Exec(r.Context(), insertQuery, "indeed", job.Title, job.Company, job.Location, job.URL)
		if err == nil {
			jobsInserted++
		}
	}

	h.json(w, ScrapeResponse{JobsScraped: jobsInserted}, http.StatusOK)
}
