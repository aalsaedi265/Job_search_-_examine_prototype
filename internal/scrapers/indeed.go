package scrapers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Job struct {
	Title    string
	Company  string
	Location string
	URL      string
}

type IndeedScraper struct {
	allocCtx context.Context
	cancel   context.CancelFunc
}

func NewIndeedScraper() (*IndeedScraper, error) {
	// Use default Chrome options - simplest approach
	allocCtx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
		)...,
	)

	return &IndeedScraper{
		allocCtx: allocCtx,
		cancel:   cancel,
	}, nil
}

func (s *IndeedScraper) Close() {
	s.cancel()
}

func (s *IndeedScraper) Scrape(keywords, location string) ([]Job, error) {
	ctx, cancel := chromedp.NewContext(s.allocCtx)
	defer cancel()

	// Set reasonable timeout (20 seconds is enough for modern Indeed pages)
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Build Indeed URL
	indeedURL := fmt.Sprintf(
		"https://www.indeed.com/jobs?q=%s&l=%s",
		url.QueryEscape(keywords),
		url.QueryEscape(location),
	)
	log.Printf("Scraping Indeed URL: %s", indeedURL)

	var jobs []Job

	// Navigate and scrape with modern Indeed selectors (2025)
	err := chromedp.Run(ctx,
		chromedp.Navigate(indeedURL),
		// Wait for job cards container to load
		chromedp.WaitVisible(`.job_seen_beacon, .tapItem`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Give Indeed time to render JavaScript

		chromedp.Evaluate(`
			// Modern Indeed (2025) uses .job_seen_beacon for job cards
			let jobCards = document.querySelectorAll('.job_seen_beacon');

			// Fallback to .tapItem if no job_seen_beacon found
			if (jobCards.length === 0) {
				jobCards = document.querySelectorAll('.tapItem');
			}

			console.log("Found " + jobCards.length + " job cards on Indeed");

			Array.from(jobCards).slice(0, 10).map(card => {
				// Modern Indeed selectors (verified 2025)
				const titleEl = card.querySelector('h2.jobTitle span') ||
				                card.querySelector('h2.jobTitle a span');

				const companyEl = card.querySelector('span.companyName');

				const locationEl = card.querySelector('div.companyLocation');

				// Link can be on the title or the card itself
				const linkEl = card.querySelector('h2.jobTitle a') ||
				               card.querySelector('a[data-jk]');

				return {
					title: titleEl ? titleEl.textContent.trim() : '',
					company: companyEl ? companyEl.textContent.trim() : '',
					location: locationEl ? locationEl.textContent.trim() : '',
					url: linkEl ? (linkEl.href || ('https://www.indeed.com' + linkEl.getAttribute('href'))) : ''
				};
			});
		`, &jobs),
	)

	if err != nil {
		log.Printf("ChromeDP error: %v", err)
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	log.Printf("Raw scraped jobs count: %d", len(jobs))

	// Filter out empty jobs
	validJobs := []Job{}
	for _, job := range jobs {
		if job.Title != "" && job.Company != "" && job.URL != "" {
			// Clean up URL (remove query params after job ID)
			if idx := strings.Index(job.URL, "?"); idx != -1 {
				job.URL = job.URL[:idx]
			}
			validJobs = append(validJobs, job)
		}
	}

	return validJobs, nil
}
