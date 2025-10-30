# Job Application Automation Tool

A full-stack application for automating job applications with a Go backend API and Svelte frontend.

## Quick Start Guide

### Prerequisites

1. **Go 1.21+** - [Download](https://go.dev/dl/)
   ```bash
   go version
   ```

2. **Node.js 18+** - [Download](https://nodejs.org/)
   ```bash
   node --version
   npm --version
   ```

3. **PostgreSQL 15+** - [Download](https://www.postgresql.org/download/)
   ```bash
   psql --version
   ```

### Setup in 5 Steps

#### 1. Database Setup
```bash
# Create the database
createdb jobapply_db

# Or using psql directly:
psql -U postgres
CREATE DATABASE jobapply_db;
\q
```

#### 2. Backend Setup
```bash
# Install Go dependencies
go mod download

# Create environment file
cp .env.example .env

# Edit .env and set your DATABASE_URL:
# DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@localhost:5432/jobapply_db?sslmode=disable
```

#### 3. Start Backend
```bash
# Run the backend (migrations run automatically)
go run cmd/api/main.go
```
Backend will be running at `http://localhost:8080`

#### 4. Start Frontend (Open New Terminal)
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies (first time only)
npm install

# Start development server
npm run dev
```
Frontend will be running at `http://localhost:5173`

#### 5. Test the Application
Open your browser to `http://localhost:5173` and create your first profile!

## Features

- **User Profile Management**: Store complete user profiles including work history, education, skills, and contact information
- **Resume Upload**: Upload and store PDF resumes (manual work history entry for accuracy)
- **Search Configuration**: Configure job search preferences and keywords
- **Job Scraping**: API-based job scraping from The Muse (500 req/day free tier)
- **Smart Caching**: 12-hour cache reduces API usage by ~90%
- **Health Monitoring**: Built-in health check endpoint for monitoring
- **PostgreSQL Database**: Robust data persistence with proper indexing
- **Embedded Migrations**: Automatic database schema setup

### Known Limitations
- The Muse API only supports broad job categories (e.g., "Software Engineering"), not specific keywords (e.g., "software engineering manager")
- Future: Migrate to Adzuna API for true keyword search

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Chi router (lightweight, composable)
- **Database**: PostgreSQL 15+ with pgx/v5
- **Job API**: The Muse API (HTTP client, no browser automation)

### Frontend
- **Framework**: Svelte 5
- **Build Tool**: Vite
- **Styling**: CSS

## Project Structure

```
Job_application/
├── backend/
│   ├── cmd/api/
│   │   └── main.go                 # Application entry point
│   ├── internal/
│   │   ├── database/
│   │   │   ├── db.go               # PostgreSQL connection + migrations
│   │   │   └── migrations/         # SQL migration files
│   │   ├── models/
│   │   │   └── models.go           # All data models
│   │   ├── handlers/
│   │   │   ├── auth.go             # Authentication handlers
│   │   │   ├── handlers.go         # HTTP handlers
│   │   │   └── scraping.go         # Job scraping with caching
│   │   ├── scrapers/
│   │   │   └── muse.go             # The Muse API client
│   │   ├── middleware/
│   │   │   └── security.go         # Security middleware
│   │   └── validation/
│   │       └── sanitize.go         # Input sanitization
│   ├── uploads/                    # Uploaded resume storage
│   ├── .env.example                # Environment template
│   ├── Makefile                    # Common commands
│   └── go.mod                      # Go dependencies
│
├── frontend/
│   ├── src/
│   │   ├── components/             # Svelte components
│   │   │   ├── ProfileForm.svelte
│   │   │   ├── SearchConfig.svelte
│   │   │   └── JobList.svelte
│   │   ├── lib/                    # Utilities and stores
│   │   └── App.svelte             # Main app component
│   ├── public/                     # Static assets
│   ├── package.json
│   └── vite.config.js
│
└── README.md                       # This file
```

**Architecture Philosophy:**
- **Lean backend**: 8 Go files organized by function - extremely simple
- **No repository layer**: Handlers talk directly to database
- **No config package**: Environment loading in main.go
- **Direct approach**: Minimal abstraction for maximum maintainability
- **No browser automation**: Simple HTTP API calls instead of ChromeDP complexity

## Configuration

All configuration is done via environment variables. Copy `.env.example` to `.env` and update the values:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | *Required* |
| `PORT` | Server port | `8080` |
| `UPLOAD_DIR` | Directory for uploaded files | `./uploads` |
| `MAX_UPLOAD_SIZE` | Max file upload size in bytes | `5242880` (5MB) |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:5173` |

## API Endpoints

### Health Check

**GET** `/health`

Returns server and database health status.

**Response:**
```json
{
  "status": "ok",
  "database": "connected",
  "time": "2025-10-06T10:00:00Z"
}
```

### Profile Management

#### Create/Update Profile

**POST** `/api/v1/profile`

Creates a new profile or updates existing one by email (upsert).

**Request Body:**
```json
{
  "full_name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1234567890",
  "address": {
    "street": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "zip_code": "94105"
  },
  "work_history": [
    {
      "company": "Tech Corp",
      "title": "Software Engineer",
      "start_date": "2020-01-01",
      "end_date": "2023-12-31",
      "description": "Developed web applications"
    }
  ],
  "education": [
    {
      "school": "University of California",
      "degree": "Bachelor of Science",
      "major": "Computer Science",
      "grad_year": 2019
    }
  ],
  "skills": ["Go", "Python", "PostgreSQL", "Docker"]
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "full_name": "John Doe",
  "email": "john.doe@example.com",
  ...
  "created_at": "2025-10-06T10:00:00Z",
  "updated_at": "2025-10-06T10:00:00Z"
}
```

#### Get Profile by ID

**GET** `/api/v1/profile/{id}`

Retrieves a user profile by UUID.

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "full_name": "John Doe",
  ...
}
```

#### Upload Resume

**POST** `/api/v1/profile/{id}/resume`

Uploads a PDF resume for a profile.

**Request:** `multipart/form-data`
- Field name: `resume`
- File type: PDF only
- Max size: 5MB (configurable)

**Response:** `200 OK`
```json
{
  "resume_url": "/uploads/550e8400-e29b-41d4-a716-446655440000.pdf",
  "message": "Resume uploaded successfully"
}
```

### Search Configuration

#### Create/Update Search Config

**POST** `/api/v1/search-config`

Creates or updates job search configuration for a user.

**Request Body:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "keywords": ["software engineer", "golang", "backend"],
  "timeframe": "24h",
  "enabled_sites": ["linkedin", "indeed", "glassdoor"]
}
```

**Response:** `201 Created`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "keywords": ["software engineer", "golang", "backend"],
  ...
}
```

#### Get Search Config by User ID

**GET** `/api/v1/search-config/{user_id}`

Retrieves search configuration for a user.

**Response:** `200 OK`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  ...
}
```

## Database Schema

The application creates four tables:

1. **user_profiles** - User information and resume data
2. **search_configs** - Job search preferences
3. **jobs** - Job listings (for future phases)
4. **applications** - Application tracking (for future phases)

Migrations run automatically on server startup.

## Development Commands

### Backend Commands

```bash
# Run the server
go run cmd/api/main.go
# or
make run

# Build binary
make build              # Creates bin/jobapply-api

# Run tests
go test -v ./...
# or
make test

# Clean build artifacts
make clean

# Install/update dependencies
go mod download
# or
make deps
```

### Frontend Commands

```bash
cd frontend

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Testing with cURL

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Create Profile
```bash
curl -X POST http://localhost:8080/api/v1/profile \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "skills": ["Go", "Python", "PostgreSQL"]
  }'
```

### 3. Get Profile (replace {id} with actual UUID)
```bash
curl http://localhost:8080/api/v1/profile/{id}
```

### 4. Upload Resume (replace {id} with actual UUID)
```bash
curl -X POST http://localhost:8080/api/v1/profile/{id}/resume \
  -F "resume=@/path/to/your/resume.pdf"
```

### 5. Create Search Config
```bash
curl -X POST http://localhost:8080/api/v1/search-config \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "{your-profile-id}",
    "keywords": ["software engineer", "golang"],
    "timeframe": "24h",
    "enabled_sites": ["linkedin", "indeed"]
  }'
```

## Error Handling

All errors return JSON with appropriate HTTP status codes:

```json
{
  "error": "Error message description"
}
```

Common status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `404` - Not Found
- `500` - Internal Server Error
- `503` - Service Unavailable (database down)

## Development Notes

- **UUID IDs**: All records use UUIDs instead of auto-incrementing integers
- **JSONB Storage**: Complex objects (address, work history, education) are stored as JSONB for flexibility
- **Connection Pooling**: PostgreSQL connection pool is configured with min 5, max 25 connections
- **CORS**: Enabled for `http://localhost:5173` by default (configurable)
- **Graceful Shutdown**: Server handles SIGTERM/SIGINT with 30-second grace period

## Roadmap

### Phase 1 (Current)
- [x] User profile management
- [x] Resume uploads
- [x] Search configuration
- [x] Basic job scraping (Indeed.com)
- [x] Frontend UI with Svelte

### Phase 2 (Planned)
- [ ] Browser automation for form filling
- [ ] Custom question handling with AI
- [ ] User approval workflow
- [ ] Application tracking dashboard

### Future Phases
- [ ] Authentication & multi-user support
- [ ] Email notifications
- [ ] Advanced job matching algorithms
- [ ] Analytics and reporting

## Troubleshooting

### Database Connection Issues
```bash
# Test PostgreSQL connection
psql "postgresql://postgres:password@localhost:5432/jobapply_db"

# Check if database exists
psql -U postgres -l

# Check if PostgreSQL is running (Windows)
Get-Service postgresql*

# Check if PostgreSQL is running (Mac/Linux)
brew services list    # if installed via Homebrew
sudo systemctl status postgresql  # if installed via apt/yum
```

### Migration Issues
```bash
# Manually run migrations
psql $DATABASE_URL -f internal/database/migrations/001_initial_schema.up.sql

# Rollback migrations
psql $DATABASE_URL -f internal/database/migrations/001_initial_schema.down.sql

# Drop and recreate database (nuclear option)
dropdb jobapply_db
createdb jobapply_db
# Then restart backend to rerun migrations
```

### Port Already in Use

Backend:
```bash
# Change PORT in .env file
PORT=8081
```

Frontend:
```bash
# Vite will automatically try the next available port (5174, 5175, etc.)
# Or manually specify in vite.config.js:
# server: { port: 3000 }
```

### CORS Errors in Frontend
```bash
# Make sure ALLOWED_ORIGINS in .env matches your frontend URL
ALLOWED_ORIGINS=http://localhost:5173

# If frontend runs on different port, update this value
```

### Resume Upload Fails
```bash
# Ensure uploads directory exists
mkdir -p uploads

# Check file permissions (Windows)
icacls uploads

# Check file permissions (Mac/Linux)
chmod 755 uploads

# Verify MAX_UPLOAD_SIZE in .env (default 5MB)
MAX_UPLOAD_SIZE=10485760  # 10MB
```

### Frontend Build Fails
```bash
cd frontend

# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
npm run dev
```