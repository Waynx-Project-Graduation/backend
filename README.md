# AI Trip Planner — Backend API

A Go-based REST API for the AI Trip Planning Engine graduation project, featuring advanced AI integration, robust database relationships, and strict data validation.

## Tech Stack

- **Go 1.21+** with **Gin** web framework
- **SQLite** with **GORM** ORM (Configured for safe concurrent writes)
- **JWT** authentication (access + refresh tokens)
- **Google Gemini AI** integration (WAYNX API for trip planning & chat)

## Key Architecture & Features

- **Robust Database Integrity:** Uses strict transaction blocks and database-level `OnDelete:CASCADE` constraints to ensure there is never any orphaned or ghost data. Uses hard deletes to save space.
- **AI JSON Sanitization:** Automatically strips markdown and cleanly parses Google Gemini responses even when the AI hallucinates formatting.
- **Security:** Hardened bcrypt password hashing (cost 12), secure JWT endpoints, and sanitized internal SQL errors so internal schemas are never leaked to the client.
- **Advanced Entities:** Supports Trip Collaboration, Budget Tracking, Place Reviews, and User Roles.

## Quick Start

### 1. Prerequisites
- Go 1.21+
- A valid Google Gemini API key

### 2. Setup

```bash
# Clone and enter backend directory
cd backend

# Copy and edit environment variables
cp .env.example .env
# Edit .env with your settings (database will be automatically created via SQLite)

# Install dependencies
go mod tidy

# Run database migrations + seed sample data
go run cmd/seed/main.go

# Start the server
go run cmd/server/main.go
```

## API Endpoints

Server runs on `http://localhost:8080`

### Auth (Public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/google` | Google OAuth |
| POST | `/api/auth/forgot-password` | Request password reset |
| POST | `/api/auth/reset-password` | Reset password with token |
| POST | `/api/auth/refresh` | Refresh token (no auth required) |

### Users (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/me` | Get current user |
| PUT | `/api/users/profile` | Update profile (name, city, avatar) |
| PUT | `/api/users/preferences` | Update travel preferences |
| PUT | `/api/users/password` | Change password |
| GET | `/api/users/stats` | Get profile statistics |
| GET | `/api/users/saved-places` | List saved/bookmarked places |

### Trips (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/trips` | Create trip (AI dynamically generates itinerary inside a DB transaction) |
| GET | `/api/trips` | List my trips |
| GET | `/api/trips/:id` | Get trip with destinations → days → activities |
| PUT | `/api/trips/:id` | Update trip metadata |
| DELETE | `/api/trips/:id` | Hard delete trip (cascades to all dependencies instantly) |
| POST | `/api/trips/:id/regenerate` | Wipe itinerary and re-generate from AI |

### Places & Chat
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/places` | List places (supports advanced search and filtering) |
| GET | `/api/places/:id` | Get place details |
| POST | `/api/places/:id/save` | Save place to favorites |
| POST | `/api/chat` | Send message to WAYNX AI Assistant |

## Advanced Database Schema

The database utilizes highly relational and normalized tables.

### The Trip Hierarchy
When the AI generates a trip, it creates records across four distinct tables wrapped in a single transaction:
```
trips
  └── trip_destinations (e.g. Cairo 3 days, Luxor 2 days)
        └── trip_days (Day 1, Day 2)
              └── trip_activities (Pyramids 4h, Museum 3h)
```

### Advanced Features & Tables
- **Trip Members (`trip_members`)**: A junction table that allows many-to-many collaboration on trips. Users can be assigned roles (e.g. owner, editor).
- **Trip Expenses (`trip_expenses`)**: Dedicated table for tracking individual expenses per trip (amount, currency, category).
- **Place Reviews (`place_reviews`)**: A junction table allowing users to rate (1-5) and comment on places.
- **User Roles (`users.role`)**: Role-based access control column serving as the foundation for an Admin Panel.

## Trip Creation — Request Payload

Trips use the same `preferences` JSON object as the user profile. The AI picks destinations automatically and strictly validates inputs.

**Request:** `POST /api/trips`
```json
{
  "start_date": "2026-05-01",
  "end_date": "2026-05-07",
  "travelers_count": 2,
  "preferences": {
    "interests": ["history", "adventure"],
    "travel_companion": "couple",
    "budget": "medium",
    "age_group": "adult",
    "crowd_preference": "quiet",
    "season": "winter"
  }
}
```

## Project Structure

```
backend/
├── cmd/
│   ├── server/main.go           # App entry point
│   └── seed/main.go             # Database seeder
├── internal/
│   ├── database/                # SQLite connection + safe AutoMigrate
│   ├── handlers/                # HTTP route handlers (sanitized errors)
│   ├── models/                  # GORM models (User, Trip, Review, Expense, Member)
│   ├── repository/              # SQL queries and DB Transaction wrappers
│   ├── services/                # Business logic + Gemini AI Markdown Sanitization
│   └── utils/                   # JWT generation, response helpers
├── .env.example
├── go.mod
└── go.sum
```
