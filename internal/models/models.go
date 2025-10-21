package models

import "time"

// Address represents a user's address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

// WorkHistory represents a user's work experience
type WorkHistory struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
	Description string `json:"description"`
}

// Education represents a user's educational background
type Education struct {
	School   string `json:"school"`
	Degree   string `json:"degree"`
	Major    string `json:"major"`
	GradYear int    `json:"grad_year"`
}

// UserProfile represents a user's complete profile
type UserProfile struct {
	ID          string        `json:"id"`
	FullName    string        `json:"full_name"`
	Email       string        `json:"email"`
	Phone       string        `json:"phone,omitempty"`
	Address     *Address      `json:"address,omitempty"`
	WorkHistory []WorkHistory `json:"work_history,omitempty"`
	Education   []Education   `json:"education,omitempty"`
	ResumeURL   *string       `json:"resume_url,omitempty"`
	Skills      []string      `json:"skills,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
