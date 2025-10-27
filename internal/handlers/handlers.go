package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/jobapply/internal/models"
	"github.com/yourusername/jobapply/internal/services"
	"github.com/yourusername/jobapply/internal/validation"
)

type Handler struct {
	db             *pgxpool.Pool
	uploadDir      string
	maxUploadSize  int64
	browserManager *services.BrowserManager
}

func New(db *pgxpool.Pool, uploadDir string, maxUploadSize int64, browserManager *services.BrowserManager) *Handler {
	return &Handler{
		db:             db,
		uploadDir:      uploadDir,
		maxUploadSize:  maxUploadSize,
		browserManager: browserManager,
	}
}

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	if err := h.db.Ping(ctx); err != nil {
		resp["status"] = "error"
		resp["database"] = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		resp["database"] = "connected"
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateProfile updates the authenticated user's profile
func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE user_profiles
		SET full_name = $1, phone = $2, address = $3, work_history = $4, education = $5, skills = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING id, full_name, email, phone, address, work_history, education, resume_url, skills, created_at, updated_at
	`

	var profile models.UserProfile
	err := h.db.QueryRow(r.Context(), query,
		req.FullName,
		req.Phone,
		toJSON(req.Address), toJSON(req.WorkHistory), toJSON(req.Education),
		req.Skills,
		userID,
	).Scan(
		&profile.ID, &profile.FullName, &profile.Email, &profile.Phone,
		scanJSON(&profile.Address), scanJSON(&profile.WorkHistory), scanJSON(&profile.Education),
		&profile.ResumeURL, &profile.Skills, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		h.error(w, fmt.Sprintf("Failed to update profile: %v", err), http.StatusInternalServerError)
		return
	}

	h.json(w, profile, http.StatusOK)
}

// GetProfile gets the authenticated user's profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.getUserProfile(r.Context(), userID)
	if err != nil {
		if err.Error() == "profile not found" {
			h.error(w, "Profile not found", http.StatusNotFound)
			return
		}
		h.error(w, fmt.Sprintf("Failed to get profile: %v", err), http.StatusInternalServerError)
		return
	}

	h.json(w, *profile, http.StatusOK)
}

// UploadResume uploads a resume file for the authenticated user with security checks
func (h *Handler) UploadResume(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Limit form parsing size to prevent memory exhaustion
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		h.error(w, "File too large or invalid request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		h.error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Double-check file size to prevent bypasses
	if header.Size > h.maxUploadSize {
		h.error(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}

	// Minimum file size check (prevent empty or tiny malicious files)
	if header.Size < 100 {
		h.error(w, "File too small to be a valid resume", http.StatusBadRequest)
		return
	}

	// Sanitize original filename to prevent path traversal
	sanitizedName := validation.SanitizeFilename(header.Filename)

	// Validate file extension using whitelist
	if !validation.ValidateFileExtension(sanitizedName, []string{".pdf"}) {
		h.error(w, "Only PDF files allowed", http.StatusBadRequest)
		return
	}

	// Read file content to verify it's actually a PDF (magic number check)
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		h.error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Check PDF magic number signature (%PDF)
	if n < 4 || !bytes.HasPrefix(buffer[:n], []byte("%PDF")) {
		h.error(w, "Invalid PDF file (file content does not match PDF format)", http.StatusBadRequest)
		return
	}

	// Reset file pointer to beginning for copying
	if _, err := file.Seek(0, 0); err != nil {
		h.error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Generate secure random filename (prevents guessing and overwrites)
	filename := fmt.Sprintf("%s.pdf", uuid.New().String())
	filePath := filepath.Join(h.uploadDir, filename)

	os.MkdirAll(h.uploadDir, 0755)

	dst, err := os.Create(filePath)
	if err != nil {
		h.error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	resumeURL := fmt.Sprintf("/uploads/%s", filename)

	result, err := h.db.Exec(r.Context(), "UPDATE user_profiles SET resume_url = $1, updated_at = NOW() WHERE id = $2", resumeURL, userID)
	if err != nil || result.RowsAffected() == 0 {
		os.Remove(filePath)
		h.error(w, "Profile not found", http.StatusNotFound)
		return
	}

	h.json(w, map[string]string{"resume_url": resumeURL, "message": "Resume uploaded successfully"}, http.StatusOK)
}

// GetJobs gets all scraped jobs
func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, title, company, location, url, scraped_at
		FROM jobs
		ORDER BY scraped_at DESC
		LIMIT 50
	`

	rows, err := h.db.Query(r.Context(), query)
	if err != nil {
		h.error(w, fmt.Sprintf("Failed to get jobs: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Job struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Company   string    `json:"company"`
		Location  string    `json:"location"`
		URL       string    `json:"url"`
		ScrapedAt time.Time `json:"scraped_at"`
	}

	jobs := []Job{}
	for rows.Next() {
		var job Job
		var location *string
		if err := rows.Scan(&job.ID, &job.Title, &job.Company, &location, &job.URL, &job.ScrapedAt); err != nil {
			continue
		}
		if location != nil {
			job.Location = *location
		}
		jobs = append(jobs, job)
	}

	h.json(w, jobs, http.StatusOK)
}

// DeleteProfile deletes the authenticated user's profile and associated data
func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get profile first to delete resume file
	profile, err := h.getUserProfile(r.Context(), userID)
	if err == nil && profile.ResumeURL != nil && *profile.ResumeURL != "" {
		// Delete the resume file if it exists
		filename := filepath.Base(*profile.ResumeURL)
		filePath := filepath.Join(h.uploadDir, filename)
		os.Remove(filePath) // Ignore errors - file might not exist
	}

	// Delete the user profile
	result, err := h.db.Exec(r.Context(), "DELETE FROM user_profiles WHERE id = $1", userID)
	if err != nil {
		h.error(w, fmt.Sprintf("Failed to delete profile: %v", err), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		h.error(w, "Profile not found", http.StatusNotFound)
		return
	}

	h.json(w, map[string]string{"message": "Profile deleted successfully"}, http.StatusOK)
}

// ValidateProfile checks if the authenticated user's profile is complete enough for job searching
func (h *Handler) ValidateProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.getUserProfile(r.Context(), userID)
	if err != nil {
		h.error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Calculate years of experience from work history
	var totalYears float64
	var missingFields []string

	// Check required fields
	if profile.FullName == "" {
		missingFields = append(missingFields, "full_name")
	}
	if profile.Email == "" {
		missingFields = append(missingFields, "email")
	}
	if len(profile.WorkHistory) == 0 {
		missingFields = append(missingFields, "work_history")
	} else {
		// Calculate years of experience
		for _, work := range profile.WorkHistory {
			if work.StartDate != "" {
				startDate, err := time.Parse("2006-01-02", work.StartDate)
				if err == nil {
					var endDate time.Time
					if work.EndDate != "" {
						endDate, err = time.Parse("2006-01-02", work.EndDate)
						if err != nil {
							endDate = time.Now()
						}
					} else {
						endDate = time.Now() // Current job
					}
					years := endDate.Sub(startDate).Hours() / 24 / 365.25
					totalYears += years
				}
			}
		}
	}

	type ValidationResponse struct {
		IsComplete        bool     `json:"is_complete"`
		YearsOfExperience float64  `json:"years_of_experience"`
		MissingFields     []string `json:"missing_fields,omitempty"`
		Message           string   `json:"message,omitempty"`
	}

	response := ValidationResponse{
		IsComplete:        len(missingFields) == 0,
		YearsOfExperience: totalYears,
		MissingFields:     missingFields,
	}

	if !response.IsComplete {
		response.Message = "Please complete your profile before searching for jobs"
	} else {
		response.Message = fmt.Sprintf("Profile complete with %.1f years of experience", totalYears)
	}

	h.json(w, response, http.StatusOK)
}

// GetApplications gets applications for the authenticated user
func (h *Handler) GetApplications(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT a.id, a.status, a.applied_at, a.filled_fields, j.title, j.company, j.url
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE a.user_id = $1
		ORDER BY a.applied_at DESC
	`

	rows, err := h.db.Query(r.Context(), query, userID)
	if err != nil {
		h.error(w, fmt.Sprintf("Failed to get applications: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Application struct {
		ID           string    `json:"id"`
		Status       string    `json:"status"`
		AppliedAt    time.Time `json:"applied_at"`
		FieldsFilled []string  `json:"fields_filled"`
		JobTitle     string    `json:"job_title"`
		Company      string    `json:"company"`
		JobURL       string    `json:"job_url"`
	}

	applications := []Application{}
	for rows.Next() {
		var app Application
		var filledFieldsJSON []byte
		if err := rows.Scan(&app.ID, &app.Status, &app.AppliedAt, &filledFieldsJSON, &app.JobTitle, &app.Company, &app.JobURL); err != nil {
			continue
		}

		// Parse filled_fields JSON
		if len(filledFieldsJSON) > 0 {
			var fieldsData map[string][]string
			if err := json.Unmarshal(filledFieldsJSON, &fieldsData); err == nil {
				if fields, ok := fieldsData["fields"]; ok {
					app.FieldsFilled = fields
				}
			}
		}

		applications = append(applications, app)
	}

	h.json(w, applications, http.StatusOK)
}

// Helper functions
func (h *Handler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) error(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// validateUUID validates a UUID string and sends error response if invalid
func (h *Handler) validateUUID(w http.ResponseWriter, id, fieldName string) bool {
	if _, err := uuid.Parse(id); err != nil {
		h.error(w, fmt.Sprintf("Invalid %s format", fieldName), http.StatusBadRequest)
		return false
	}
	return true
}

// getUserProfile fetches a user profile by ID from the database
func (h *Handler) getUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	query := `
		SELECT id, full_name, email, phone, address, work_history, education, resume_url, skills, created_at, updated_at
		FROM user_profiles WHERE id = $1
	`

	var profile models.UserProfile
	err := h.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.FullName, &profile.Email, &profile.Phone,
		scanJSON(&profile.Address), scanJSON(&profile.WorkHistory), scanJSON(&profile.Education),
		&profile.ResumeURL, &profile.Skills, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("profile not found")
		}
		return nil, err
	}

	return &profile, nil
}

func toJSON(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func scanJSON(v interface{}) interface{} {
	return &jsonScanner{v: v}
}

type jsonScanner struct {
	v interface{}
}

func (s *jsonScanner) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into json", src)
	}

	return json.Unmarshal(b, s.v)
}
