package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	"github.com/yourusername/jobapply/internal/scrapers"
)

// ResumeRequest represents the request body for resuming an application
type ResumeRequest struct {
	Answers map[string]string `json:"answers"` // question ID -> answer
}

// ResumeResponse represents the response after resuming
type ResumeResponse struct {
	Status        string                      `json:"status"` // "submitted", "paused", "failed"
	Message       string                      `json:"message"`
	Questions     []scrapers.CustomQuestion   `json:"questions,omitempty"` // If more questions found
	FieldsFilled  []string                    `json:"fields_filled"`
	Errors        []string                    `json:"errors"`
}

// ResumeApplication handles POST /api/v1/apply/:application_id/resume
func (h *Handler) ResumeApplication(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "application_id")
	if !h.validateUUID(w, applicationID, "application_id") {
		return
	}

	var req ResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Resuming application %s with %d answers", applicationID, len(req.Answers))

	// Get application from database
	var userID, jobID, status, currentURL string
	var customQuestionsJSON []byte
	query := `
		SELECT user_id, job_id, status, custom_questions, current_url
		FROM applications WHERE id = $1
	`
	err := h.db.QueryRow(r.Context(), query, applicationID).Scan(
		&userID, &jobID, &status, &customQuestionsJSON, &currentURL,
	)
	if err != nil {
		log.Printf("Application not found: %v", err)
		h.error(w, "Application not found", http.StatusNotFound)
		return
	}

	if status != "paused" {
		h.error(w, fmt.Sprintf("Application is not paused (current status: %s)", status), http.StatusBadRequest)
		return
	}

	// Retrieve browser context
	browserCtx, exists := h.browserManager.Get(applicationID)
	if !exists {
		log.Printf("Browser context not found for application %s", applicationID)
		h.error(w, "Browser session expired. Please start a new application.", http.StatusGone)

		// Update status to timeout
		h.db.Exec(r.Context(), "UPDATE applications SET status = 'timeout' WHERE id = $1", applicationID)
		return
	}

	// Parse stored questions
	storedQuestions, err := scrapers.JSONToQuestions(customQuestionsJSON)
	if err != nil {
		log.Printf("Failed to parse stored questions: %v", err)
		h.error(w, "Failed to parse questions", http.StatusInternalServerError)
		return
	}

	// Fill in answers
	fieldsFilled := []string{}
	errors := []string{}

	for _, question := range storedQuestions {
		answer, hasAnswer := req.Answers[question.ID]
		if !hasAnswer {
			if question.Required {
				errors = append(errors, fmt.Sprintf("Missing required answer for: %s", question.Label))
			}
			continue
		}

		log.Printf("Filling answer for question '%s': %s", question.Label, answer)
		err := scrapers.FillAnswer(browserCtx.Ctx, question, answer)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to fill '%s': %v", question.Label, err)
			log.Printf(errMsg)
			errors = append(errors, errMsg)
		} else {
			fieldsFilled = append(fieldsFilled, question.Label)
		}
	}

	// Save user answers to database
	answersJSON := toJSON(req.Answers)
	h.db.Exec(r.Context(), "UPDATE applications SET user_answers = $1 WHERE id = $2", answersJSON, applicationID)

	// Check if there are validation errors from filling
	if len(errors) > 0 {
		log.Printf("Errors filling answers, returning to user")
		h.json(w, ResumeResponse{
			Status:       "paused",
			Message:      "Some fields could not be filled",
			Questions:    storedQuestions,
			FieldsFilled: fieldsFilled,
			Errors:       errors,
		}, http.StatusOK)
		return
	}

	// Wait a moment for form processing
	chromedp.Run(browserCtx.Ctx, chromedp.Sleep(1*time.Second))

	// Check for "Next" button (multi-page form)
	hasNext, _ := scrapers.HasNextButton(browserCtx.Ctx)
	if hasNext {
		log.Printf("Found 'Next' button, clicking to go to next page")

		err := scrapers.ClickNextButton(browserCtx.Ctx)
		if err != nil {
			log.Printf("Failed to click next button: %v", err)
			errors = append(errors, "Failed to click next button")
		} else {
			// Wait for next page to load
			chromedp.Run(browserCtx.Ctx, chromedp.Sleep(2*time.Second))

			// Detect questions on new page
			newQuestions, err := scrapers.DetectCustomQuestions(browserCtx.Ctx)
			if err != nil {
				log.Printf("Failed to detect questions on new page: %v", err)
			}

			if len(newQuestions) > 0 {
				log.Printf("Found %d more questions on next page, staying paused", len(newQuestions))

				// Save new questions and stay paused
				newQuestionsJSON := scrapers.QuestionToJSON(newQuestions)
				h.db.Exec(r.Context(), "UPDATE applications SET custom_questions = $1 WHERE id = $2", newQuestionsJSON, applicationID)

				h.json(w, ResumeResponse{
					Status:       "paused",
					Message:      "More questions found on next page",
					Questions:    newQuestions,
					FieldsFilled: fieldsFilled,
					Errors:       errors,
				}, http.StatusOK)
				return
			}
		}
	}

	// No more questions - submit the application
	log.Printf("No more questions, attempting to submit application")

	submitClicked := false
	submitSelectors := []string{
		`button[type="submit"]`,
		`button:contains("Submit")`,
		`button:contains("Submit application")`,
		`button:contains("Apply")`,
		`input[type="submit"]`,
		`button[id*="submit"]`,
		`button[class*="submit"]`,
	}

	for _, selector := range submitSelectors {
		err = chromedp.Run(browserCtx.Ctx,
			chromedp.Click(selector, chromedp.NodeVisible),
		)
		if err == nil {
			log.Printf("Clicked submit button with selector: %s", selector)
			submitClicked = true
			chromedp.Run(browserCtx.Ctx, chromedp.Sleep(3*time.Second)) // Wait for submission
			break
		}
	}

	applicationStatus := "failed"

	if submitClicked {
		log.Printf("Application submitted successfully")
		applicationStatus = "submitted"
	} else {
		log.Printf("Could not find submit button")
		errors = append(errors, "Could not find submit button")
	}

	message := "Application submitted successfully"
	if applicationStatus == "failed" {
		message = "Application failed"
	}

	// Update database
	updateQuery := `
		UPDATE applications
		SET status = $1, applied_at = $2, error_log = $3
		WHERE id = $4
	`
	errorLog := ""
	if len(errors) > 0 {
		errorLog = fmt.Sprintf("%v", errors)
	}
	h.db.Exec(r.Context(), updateQuery, applicationStatus, time.Now(), errorLog, applicationID)

	// Clean up browser context
	h.browserManager.Remove(applicationID)

	h.json(w, ResumeResponse{
		Status:       applicationStatus,
		Message:      message,
		FieldsFilled: fieldsFilled,
		Errors:       errors,
	}, http.StatusOK)
}

// CancelApplication handles DELETE /api/v1/apply/:application_id
func (h *Handler) CancelApplication(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "application_id")
	if !h.validateUUID(w, applicationID, "application_id") {
		return
	}

	log.Printf("Canceling application %s", applicationID)

	// Remove browser context
	h.browserManager.Remove(applicationID)

	// Update database
	h.db.Exec(r.Context(), "UPDATE applications SET status = 'cancelled' WHERE id = $1", applicationID)

	h.json(w, map[string]string{
		"message": "Application cancelled",
		"status":  "cancelled",
	}, http.StatusOK)
}

// GetApplicationStatus handles GET /api/v1/apply/:application_id/status
func (h *Handler) GetApplicationStatus(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "application_id")
	if !h.validateUUID(w, applicationID, "application_id") {
		return
	}

	var status string
	var customQuestionsJSON []byte
	var pausedAt *time.Time

	query := `
		SELECT status, custom_questions, paused_at
		FROM applications WHERE id = $1
	`
	err := h.db.QueryRow(r.Context(), query, applicationID).Scan(&status, &customQuestionsJSON, &pausedAt)
	if err != nil {
		h.error(w, "Application not found", http.StatusNotFound)
		return
	}

	// Parse questions if paused
	var questions []scrapers.CustomQuestion
	if status == "paused" && len(customQuestionsJSON) > 0 {
		questions, _ = scrapers.JSONToQuestions(customQuestionsJSON)
	}

	// Check if browser context still exists
	_, hasContext := h.browserManager.Get(applicationID)

	response := map[string]interface{}{
		"status":       status,
		"paused_at":    pausedAt,
		"has_context":  hasContext,
		"questions":    questions,
	}

	h.json(w, response, http.StatusOK)
}
