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

	// Set timeout (increased for Windows/Chrome startup)
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Build Indeed URL
	indeedURL := fmt.Sprintf(
		"https://www.indeed.com/jobs?q=%s&l=%s",
		url.QueryEscape(keywords),
		url.QueryEscape(location),
	)
	log.Printf("Scraping Indeed URL: %s", indeedURL)

	var jobs []Job

	// Navigate and scrape
	err := chromedp.Run(ctx,
		chromedp.Navigate(indeedURL),
		chromedp.WaitVisible(`div[id="mosaic-provider-jobcards"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for dynamic content

		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('.job_seen_beacon, .tapItem')).slice(0, 10).map(card => {
				const titleEl = card.querySelector('.jobTitle span[title], .jobTitle span');
				const companyEl = card.querySelector('.companyName');
				const locationEl = card.querySelector('.companyLocation');
				const linkEl = card.querySelector('a.jcs-JobTitle');

				return {
					title: titleEl ? titleEl.textContent.trim() : '',
					company: companyEl ? companyEl.textContent.trim() : '',
					location: locationEl ? locationEl.textContent.trim() : '',
					url: linkEl ? 'https://www.indeed.com' + linkEl.getAttribute('href') : ''
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
