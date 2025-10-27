package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/yourusername/jobapply/internal/database"
	"github.com/yourusername/jobapply/internal/handlers"
	"github.com/yourusername/jobapply/internal/middleware"
	"github.com/yourusername/jobapply/internal/services"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Get config from env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := getEnv("PORT", "8080")
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	maxUploadSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "5242880"), 10, 64)

	// Connect to database and run migrations
	ctx := context.Background()
	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database successfully")

	// Initialize browser manager for Phase 4 pause/resume (15 minute timeout)
	browserManager := services.NewBrowserManager(15 * time.Minute)
	defer browserManager.Shutdown()
	log.Println("Browser manager initialized")

	// Create handlers
	h := handlers.New(db, uploadDir, maxUploadSize, browserManager)

	// Setup router
	r := chi.NewRouter()

	// Security Middleware - Order matters!
	// 1. Security headers first to protect all responses
	r.Use(middleware.SecurityHeaders)

	// 2. Rate limiting to prevent DDoS (60 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(60)
	r.Use(rateLimiter.Middleware)

	// 3. Request size limiting to prevent memory exhaustion (10MB max)
	r.Use(middleware.MaxBytesMiddleware(10 * 1024 * 1024))

	// 4. Logging for audit trail
	r.Use(loggerMiddleware)

	// 5. CORS - allow frontend to communicate
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/health", h.Health)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		r.Post("/auth/signup", h.Signup)
		r.Post("/auth/login", h.Login)

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware)

			r.Get("/auth/me", h.GetMe)
			r.Put("/auth/password", h.ChangePassword)
			r.Put("/auth/email", h.UpdateEmail)
			r.Post("/profile", h.CreateProfile)
			r.Get("/profile", h.GetProfile)
			r.Delete("/profile", h.DeleteProfile)
			r.Get("/profile/validate", h.ValidateProfile)
			r.Post("/profile/resume", h.UploadResume)
			r.Post("/scrape", h.ScrapeJobs)
			r.Get("/jobs", h.GetJobs)
			r.Post("/apply", h.ApplyToJob)
			r.Get("/applications", h.GetApplications)

			// Phase 4: Pause/Resume endpoints
			r.Post("/apply/{application_id}/resume", h.ResumeApplication)
			r.Delete("/apply/{application_id}", h.CancelApplication)
			r.Get("/apply/{application_id}/status", h.GetApplicationStatus)
		})
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second, // Scraping should complete within 20s
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Simple logging middleware
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
