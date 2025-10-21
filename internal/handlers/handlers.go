package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/jobapply/internal/models"
	"github.com/yourusername/jobapply/internal/services"
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

// CreateProfile creates or updates a user profile
func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FullName == "" || req.Email == "" {
		h.error(w, "full_name and email are required", http.StatusBadRequest)
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		h.error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO user_profiles (full_name, email, phone, address, work_history, education, skills)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			work_history = EXCLUDED.work_history,
			education = EXCLUDED.education,
			skills = EXCLUDED.skills,
			updated_at = NOW()
		RETURNING id, full_name, email, phone, address, work_history, education, resume_url, skills, created_at, updated_at
	`

	var profile models.UserProfile
	err := h.db.QueryRow(r.Context(), query,
		req.FullName, req.Email, req.Phone,
		toJSON(req.Address), toJSON(req.WorkHistory), toJSON(req.Education),
		req.Skills,
	).Scan(
		&profile.ID, &profile.FullName, &profile.Email, &profile.Phone,
		scanJSON(&profile.Address), scanJSON(&profile.WorkHistory), scanJSON(&profile.Education),
		&profile.ResumeURL, &profile.Skills, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		h.error(w, fmt.Sprintf("Failed to create profile: %v", err), http.StatusInternalServerError)
		return
	}

	h.json(w, profile, http.StatusCreated)
}

// GetProfile gets a profile by ID
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.validateUUID(w, id, "ID") {
		return
	}

	profile, err := h.getUserProfile(r.Context(), id)
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

// UploadResume uploads a resume file
func (h *Handler) UploadResume(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.validateUUID(w, id, "ID") {
		return
	}

	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		h.error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		h.error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > h.maxUploadSize {
		h.error(w, "File too large", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		h.error(w, "Only PDF files allowed", http.StatusBadRequest)
		return
	}

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

	result, err := h.db.Exec(r.Context(), "UPDATE user_profiles SET resume_url = $1, updated_at = NOW() WHERE id = $2", resumeURL, id)
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

// GetApplications gets applications for a user
func (h *Handler) GetApplications(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if !h.validateUUID(w, userID, "user_id") {
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
