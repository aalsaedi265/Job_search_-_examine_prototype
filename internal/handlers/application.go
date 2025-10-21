package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/yourusername/jobapply/internal/scrapers"
)

// ApplyRequest represents the request body for job application
type ApplyRequest struct {
	JobID  string `json:"job_id"`
	UserID string `json:"user_id"`
}

// ApplyResponse represents the response after applying to a job
type ApplyResponse struct {
	ApplicationID string                    `json:"application_id"`
	Status        string                    `json:"status"` // "submitted", "paused", "failed"
	Message       string                    `json:"message"`
	FieldsFilled  []string                  `json:"fields_filled"`
	Errors        []string                  `json:"errors"`
	Questions     []scrapers.CustomQuestion `json:"questions,omitempty"` // If paused
}

// ApplyToJob handles POST /api/v1/apply - auto-fills and submits a job application
func (h *Handler) ApplyToJob(w http.ResponseWriter, r *http.Request) {
	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate UUIDs
	if !h.validateUUID(w, req.JobID, "job_id") {
		return
	}
	if !h.validateUUID(w, req.UserID, "user_id") {
		return
	}

	log.Printf("Starting application process for job_id=%s, user_id=%s", req.JobID, req.UserID)

	// Fetch job from database
	var jobURL, jobTitle string
	err := h.db.QueryRow(r.Context(), "SELECT url, title FROM jobs WHERE id = $1", req.JobID).Scan(&jobURL, &jobTitle)
	if err != nil {
		log.Printf("Job not found: %v", err)
		h.error(w, "Job not found", http.StatusNotFound)
		return
	}
	log.Printf("Found job: %s at %s", jobTitle, jobURL)

	// Fetch user profile from database
	profile, err := h.getUserProfile(r.Context(), req.UserID)
	if err != nil {
		log.Printf("User profile not found: %v", err)
		h.error(w, "User profile not found", http.StatusNotFound)
		return
	}
	log.Printf("Found user profile: %s (%s)", profile.FullName, profile.Email)

	// Parse name into first/last
	firstName, lastName := parseName(profile.FullName)

	// Get absolute resume path
	var resumePath string
	if profile.ResumeURL != nil && *profile.ResumeURL != "" {
		resumePath = filepath.Join(h.uploadDir, filepath.Base(*profile.ResumeURL))
		// Convert to absolute path
		if absPath, err := filepath.Abs(resumePath); err == nil {
			resumePath = absPath
		}
		log.Printf("Resume path: %s", resumePath)
	}

	// Initialize response tracking
	fieldsFilled := []string{}
	errors := []string{}
	applicationStatus := "failed"

	// Set up chromedp (non-headless for debugging)
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false), // Non-headless for debugging
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)...,
	)

	ctx, cancel := chromedp.NewContext(allocCtx)

	// Don't defer here - we'll manage cleanup manually based on pause/submit
	var cleanupDone bool
	defer func() {
		if !cleanupDone {
			cancel()
			allocCancel()
		}
	}()

	// Set timeout for entire operation (5 minutes for Phase 4)
	ctx, timeoutCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer timeoutCancel()

	// Perform browser automation
	log.Printf("Navigating to job: %s", jobURL)
	err = chromedp.Run(ctx,
		chromedp.Navigate(jobURL),
		chromedp.Sleep(3*time.Second), // Wait for page load
	)

	if err != nil {
		errMsg := fmt.Sprintf("Navigation failed: %v", err)
		log.Printf(errMsg)
		errors = append(errors, errMsg)
	} else {
		// Try to click Apply button
		log.Printf("Looking for Apply button...")
		applyClicked := false

		// Try multiple selectors for Apply button
		applySelectors := []string{
			`button:contains("Apply now")`,
			`button:contains("Apply")`,
			`a:contains("Apply now")`,
			`a:contains("Apply")`,
			`button[id*="apply"]`,
			`a[id*="apply"]`,
			`.jobsearch-IndeedApplyButton`,
			`.ia-IndeedApplyButton`,
		}

		for _, selector := range applySelectors {
			err = chromedp.Run(ctx,
				chromedp.Click(selector, chromedp.NodeVisible),
			)
			if err == nil {
				log.Printf("Clicked Apply button with selector: %s", selector)
				applyClicked = true
				chromedp.Run(ctx, chromedp.Sleep(2*time.Second)) // Wait for form
				break
			}
		}

		if !applyClicked {
			log.Printf("Could not find Apply button")
			errors = append(errors, "Could not find Apply button")
		} else {
			// Fill form fields
			log.Printf("Filling form fields...")

			// First Name
			if firstName != "" {
				if err := fillField(ctx, []string{
					`input[name*="firstName"]`,
					`input[id*="firstName"]`,
					`input[name*="first-name"]`,
					`input[id*="first-name"]`,
				}, firstName); err == nil {
					log.Printf("Filled field: firstName with value: %s", firstName)
					fieldsFilled = append(fieldsFilled, "firstName")
				} else {
					log.Printf("Could not find field: firstName")
				}
			}

			// Last Name
			if lastName != "" {
				if err := fillField(ctx, []string{
					`input[name*="lastName"]`,
					`input[id*="lastName"]`,
					`input[name*="last-name"]`,
					`input[id*="last-name"]`,
				}, lastName); err == nil {
					log.Printf("Filled field: lastName with value: %s", lastName)
					fieldsFilled = append(fieldsFilled, "lastName")
				} else {
					log.Printf("Could not find field: lastName")
				}
			}

			// Email
			if profile.Email != "" {
				if err := fillField(ctx, []string{
					`input[type="email"]`,
					`input[name*="email"]`,
					`input[id*="email"]`,
				}, profile.Email); err == nil {
					log.Printf("Filled field: email with value: %s", profile.Email)
					fieldsFilled = append(fieldsFilled, "email")
				} else {
					log.Printf("Could not find field: email")
				}
			}

			// Phone
			if profile.Phone != "" {
				if err := fillField(ctx, []string{
					`input[type="tel"]`,
					`input[name*="phone"]`,
					`input[id*="phone"]`,
				}, profile.Phone); err == nil {
					log.Printf("Filled field: phone with value: %s", profile.Phone)
					fieldsFilled = append(fieldsFilled, "phone")
				} else {
					log.Printf("Could not find field: phone")
				}
			}

			// City/Location (if address exists)
			if profile.Address != nil && profile.Address.City != "" {
				if err := fillField(ctx, []string{
					`input[name*="city"]`,
					`input[id*="city"]`,
					`input[name*="location"]`,
				}, profile.Address.City); err == nil {
					log.Printf("Filled field: city with value: %s", profile.Address.City)
					fieldsFilled = append(fieldsFilled, "city")
				} else {
					log.Printf("Could not find field: city")
				}
			}

			// Upload resume
			if resumePath != "" {
				log.Printf("Attempting to upload resume: %s", resumePath)
				if err := uploadResume(ctx, resumePath); err == nil {
					log.Printf("Resume uploaded successfully")
					fieldsFilled = append(fieldsFilled, "resume")
				} else {
					log.Printf("Could not upload resume: %v", err)
					errors = append(errors, fmt.Sprintf("Resume upload failed: %v", err))
				}
			}

			// PHASE 4: Detect custom questions
			log.Printf("Detecting custom screening questions...")
			customQuestions, err := scrapers.DetectCustomQuestions(ctx)
			if err != nil {
				log.Printf("Error detecting questions: %v", err)
			}

			if len(customQuestions) > 0 {
				log.Printf("Found %d custom questions - PAUSING application", len(customQuestions))
				applicationStatus = "paused"

				// Get current URL for resume reference
				var currentURL string
				chromedp.Run(ctx, chromedp.Location(&currentURL))

				// Save application to database with paused status
				var applicationID string
				insertQuery := `
					INSERT INTO applications (user_id, job_id, status, filled_fields, custom_questions, paused_at, current_url, error_log)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					RETURNING id
				`
				filledFieldsJSON := toJSON(map[string][]string{"fields": fieldsFilled})
				questionsJSON := scrapers.QuestionToJSON(customQuestions)
				errorLogText := strings.Join(errors, "; ")

				err = h.db.QueryRow(r.Context(), insertQuery,
					req.UserID, req.JobID, applicationStatus, filledFieldsJSON, questionsJSON, time.Now(), currentURL, errorLogText,
				).Scan(&applicationID)

				if err != nil {
					log.Printf("Failed to save paused application: %v", err)
					h.error(w, fmt.Sprintf("Failed to save application: %v", err), http.StatusInternalServerError)
					return
				}

				// Store browser context for later resumption - DON'T cancel it!
				h.browserManager.Store(applicationID, ctx, cancel, allocCtx, allocCancel)
				cleanupDone = true // Browser manager will handle cleanup

				log.Printf("Application paused with ID: %s, browser context stored", applicationID)

				// Return paused response with questions
				response := ApplyResponse{
					ApplicationID: applicationID,
					Status:        "paused",
					Message:       fmt.Sprintf("Application paused - %d custom questions need answers", len(customQuestions)),
					FieldsFilled:  fieldsFilled,
					Errors:        errors,
					Questions:     customQuestions,
				}

				h.json(w, response, http.StatusOK)
				return // Don't continue with submission
			}

			// No custom questions found - proceed with submission
			log.Printf("No custom questions found, proceeding with submission...")

			// Submit application
			log.Printf("Looking for Submit button...")
			submitClicked := false

			// Try multiple selectors for Submit button
			submitSelectors := []string{
				`button[type="submit"]`,
				`button:contains("Submit")`,
				`button:contains("Submit application")`,
				`button:contains("Continue")`,
				`input[type="submit"]`,
				`button[id*="submit"]`,
			}

			for _, selector := range submitSelectors {
				err = chromedp.Run(ctx,
					chromedp.Click(selector, chromedp.NodeVisible),
				)
				if err == nil {
					log.Printf("Clicked Submit button with selector: %s", selector)
					submitClicked = true
					chromedp.Run(ctx, chromedp.Sleep(3*time.Second)) // Wait for submission
					break
				}
			}

			if submitClicked {
				log.Printf("Application submitted successfully")
				applicationStatus = "submitted"
			} else {
				log.Printf("Could not find Submit button")
				errors = append(errors, "Could not find Submit button")
			}
		}
	}

	// Save application to database
	log.Printf("Saving application to database with status: %s", applicationStatus)
	var applicationID string
	insertQuery := `
		INSERT INTO applications (user_id, job_id, status, filled_fields, applied_at, error_log)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	filledFieldsJSON := toJSON(map[string][]string{"fields": fieldsFilled})
	errorLogText := strings.Join(errors, "; ")
	appliedAt := time.Now()

	err = h.db.QueryRow(r.Context(), insertQuery,
		req.UserID, req.JobID, applicationStatus, filledFieldsJSON, appliedAt, errorLogText,
	).Scan(&applicationID)

	if err != nil {
		log.Printf("Failed to save application: %v", err)
		h.error(w, fmt.Sprintf("Failed to save application: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	message := "Application submitted successfully"
	if applicationStatus == "failed" {
		message = "Application failed"
	}

	response := ApplyResponse{
		ApplicationID: applicationID,
		Status:        applicationStatus,
		Message:       message,
		FieldsFilled:  fieldsFilled,
		Errors:        errors,
	}

	log.Printf("Application process completed: %s", applicationID)
	h.json(w, response, http.StatusOK)
}

// Helper function to fill a form field with multiple selector attempts
func fillField(ctx context.Context, selectors []string, value string) error {
	for _, selector := range selectors {
		err := chromedp.Run(ctx,
			chromedp.SendKeys(selector, value, chromedp.NodeVisible),
		)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("field not found with any selector")
}

// Helper function to upload resume
func uploadResume(ctx context.Context, filePath string) error {
	// Try multiple selectors for file input
	fileSelectors := []string{
		`input[type="file"]`,
		`input[name*="resume"]`,
		`input[name*="cv"]`,
		`input[id*="resume"]`,
		`input[id*="cv"]`,
	}

	for _, selector := range fileSelectors {
		err := chromedp.Run(ctx,
			chromedp.SendKeys(selector, filePath, chromedp.NodeVisible),
		)
		if err == nil {
			chromedp.Run(ctx, chromedp.Sleep(2*time.Second)) // Wait for upload
			return nil
		}
	}
	return fmt.Errorf("file input not found")
}

// Helper function to parse full name into first and last
func parseName(fullName string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}
