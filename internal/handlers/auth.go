package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/jobapply/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

const jwtSecret = "your-secret-key-change-in-production" // TODO: Move to env var

type SignupRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateEmailRequest struct {
	NewEmail string `json:"new_email"`
}

// Signup creates a new user account
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and sanitize input to prevent XSS and injection attacks
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		h.error(w, "full_name, email, and password are required", http.StatusBadRequest)
		return
	}

	// Sanitize full name (remove HTML, limit length)
	req.FullName = validation.SanitizeString(req.FullName, 100)
	if req.FullName == "" {
		h.error(w, "Invalid full name", http.StatusBadRequest)
		return
	}

	// Validate email format using regex to prevent injection
	if !validation.ValidateEmail(req.Email) {
		h.error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password strength (6+ chars, must have letter and number)
	if !validation.ValidatePassword(req.Password) {
		h.error(w, "Password must be 6-128 characters with at least one letter and one number", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user
	query := `
		INSERT INTO user_profiles (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, full_name, email, created_at
	`

	var userID, fullName, email string
	var createdAt time.Time
	err = h.db.QueryRow(r.Context(), query, req.FullName, req.Email, string(passwordHash)).
		Scan(&userID, &fullName, &email, &createdAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			h.error(w, "Email already registered", http.StatusConflict)
			return
		}
		h.error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := generateJWT(userID, email)
	if err != nil {
		h.error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	h.json(w, AuthResponse{
		Token:  token,
		UserID: userID,
		Email:  email,
		Name:   fullName,
	}, http.StatusCreated)
}

// Login authenticates a user
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input to prevent injection attacks
	if req.Email == "" || req.Password == "" {
		h.error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	// Validate email format
	if !validation.ValidateEmail(req.Email) {
		h.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Check password length to prevent DoS with huge passwords
	if len(req.Password) > 128 {
		h.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Get user from database
	query := `
		SELECT id, full_name, email, password_hash
		FROM user_profiles
		WHERE email = $1
	`

	var userID, fullName, email, passwordHash string
	err := h.db.QueryRow(r.Context(), query, req.Email).
		Scan(&userID, &fullName, &email, &passwordHash)

	if err != nil {
		h.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		h.error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateJWT(userID, email)
	if err != nil {
		h.error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	h.json(w, AuthResponse{
		Token:  token,
		UserID: userID,
		Email:  email,
		Name:   fullName,
	}, http.StatusOK)
}

// GetMe returns the current authenticated user's profile
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.getUserProfile(r.Context(), userID)
	if err != nil {
		h.error(w, "User not found", http.StatusNotFound)
		return
	}

	h.json(w, *profile, http.StatusOK)
}

// generateJWT creates a new JWT token for a user
func generateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, `{"error":"Invalid user ID in token"}`, http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ChangePassword allows authenticated users to change their password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.CurrentPassword == "" || req.NewPassword == "" {
		h.error(w, "current_password and new_password are required", http.StatusBadRequest)
		return
	}

	// Validate new password strength
	if !validation.ValidatePassword(req.NewPassword) {
		h.error(w, "New password must be 6-128 characters with at least one letter and one number", http.StatusBadRequest)
		return
	}

	// Get current password hash
	var currentHash string
	err := h.db.QueryRow(r.Context(), "SELECT password_hash FROM user_profiles WHERE id = $1", userID).
		Scan(&currentHash)
	if err != nil {
		h.error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		h.error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.error(w, "Failed to process new password", http.StatusInternalServerError)
		return
	}

	// Update password
	_, err = h.db.Exec(r.Context(), "UPDATE user_profiles SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(newHash), userID)
	if err != nil {
		h.error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{"message": "Password changed successfully"}, http.StatusOK)
}

// UpdateEmail allows authenticated users to change their email
func (h *Handler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == "" {
		h.error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate email format
	if !validation.ValidateEmail(req.NewEmail) {
		h.error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Update email (will fail if email already exists due to unique constraint)
	result, err := h.db.Exec(r.Context(), "UPDATE user_profiles SET email = $1, updated_at = NOW() WHERE id = $2", req.NewEmail, userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			h.error(w, "Email already in use", http.StatusConflict)
			return
		}
		h.error(w, fmt.Sprintf("Failed to update email: %v", err), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		h.error(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate new JWT with updated email
	token, err := generateJWT(userID, req.NewEmail)
	if err != nil {
		h.error(w, "Failed to generate new token", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{
		"message": "Email updated successfully",
		"token":   token,
	}, http.StatusOK)
}

// getUserIDFromContext extracts the user ID from the request context
func getUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value("user_id").(string)
	return userID
}
