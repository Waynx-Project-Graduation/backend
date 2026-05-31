# AI Trip Planner — Backend API

A Go-based REST API for the AI Trip Planning Engine graduation project.

## Tech Stack

- **Go 1.21+** with **Gin** web framework
- **SQLite** with **GORM** ORM
- **JWT** authentication (access + refresh tokens)
- **Google Gemini AI** integration for trip planning & chat

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

### 3. API Endpoints

Server runs on `http://localhost:8080`

#### Auth (Public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/google` | Google OAuth |
| POST | `/api/auth/forgot-password` | Request password reset (returns token in dev mode) |
| POST | `/api/auth/reset-password` | Reset password with token |
| POST | `/api/auth/refresh` | Refresh token (no auth required) |

#### Auth (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/me` | Get current user |
| POST | `/api/auth/logout` | Logout |

#### Users (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/api/users/profile` | Update profile (name, city, avatar) |
| PUT | `/api/users/preferences` | Update travel preferences (interests, budget, etc.) |
| PUT | `/api/users/password` | Change password |
| PUT | `/api/users/avatar` | Update avatar URL |
| GET | `/api/users/stats` | Get profile statistics |
| GET | `/api/users/saved-places` | List saved/bookmarked places |

#### Trips (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/trips` | Create trip (AI picks destinations based on preferences) |
| GET | `/api/trips` | List my trips |
| GET | `/api/trips/:id` | Get trip with destinations → days → activities |
| PUT | `/api/trips/:id` | Update trip metadata |
| DELETE | `/api/trips/:id` | Delete trip (cascades to all destinations/days/activities) |
| POST | `/api/trips/:id/regenerate` | Wipe itinerary and re-generate from AI |
| PUT | `/api/trips/:id/activities/:actId` | Edit a single activity |
| DELETE | `/api/trips/:id/activities/:actId` | Remove a single activity |

#### Places (Public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/places` | List places (advanced filters) |
| GET | `/api/places/popular` | Popular places |
| GET | `/api/places/search?q=` | Search places |
| GET | `/api/places/categories` | List all categories |
| GET | `/api/places/trending` | Trending search terms |
| GET | `/api/places/:id` | Get place details |

#### Places — Favorites (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/places/:id/save` | Save place to favorites |
| DELETE | `/api/places/:id/save` | Remove from favorites |

#### Chat — Ask WAYNX (Protected 🔒)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/chat` | Send message, get AI response |
| GET | `/api/chat/history` | List chat sessions |
| GET | `/api/chat/:id` | Get chat session with messages |
| DELETE | `/api/chat/:id` | Delete chat session |

### 4. Advanced Place Filtering

The `GET /api/places` endpoint supports advanced filters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `city` | string | Single city filter (backward compatible) |
| `cities[]` | string[] | Multi-city filter |
| `category` | string | Category filter |
| `budget_level` | string | Comma-separated: "low,medium,high" |
| `budget_level[]` | string[] | Array notation: budget_level[]=low&budget_level[]=medium |
| `best_season` | string | Season: "winter", "summer", "spring", "autumn", "any" |
| `crowd_level` | string | Crowd level: "quiet", "moderate", "crowded" |
| `suitable_for`| string | Audience: "family", "couple", "solo", "friends" |
| `suitable_age`| string | Age group: "kid", "teen", "adult", "senior" |
| `sort_by` | string | "rating" (default), "name", "duration_asc", "duration_desc" |
| `search` | string | Search within results |
| `page` | int | Page number (default: 1) |
| `per_page` | int | Items per page (default: 10, max: 50) |

### 5. Trip Creation — Request & Response

Trips use the same `preferences` JSON object as the user profile. The AI picks destinations automatically.

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

**Response:** The trip includes auto-generated multi-city destinations:
```json
{
  "status": "planned",
  "destinations": [
    {
      "city": "Cairo",
      "days_allocated": 3,
      "trip_days": [
        { "day_number": 1, "activities": [...] }
      ]
    }
  ]
}
```

> If the AI service is not running or the API key is invalid, trips are saved as `"draft"` with no destinations.

### 6. Password Reset Flow (Dev Mode)

The password reset uses an in-memory token store for development:

1. **Request reset**: `POST /api/auth/forgot-password` with `{"email": "..."}` — returns a `reset_token` in the response.
2. **Reset password**: `POST /api/auth/reset-password` with `{"token": "...", "new_password": "..."}` — token is single-use and expires after 1 hour.

> In production, the token would be sent via email and not returned in the API response.

### 7. User Preferences

The `PUT /api/users/preferences` endpoint stores defaults that can pre-fill the trip form:

```json
{
  "interests": ["history", "adventure"],
  "travel_companion": "couple",
  "budget": "medium",
  "age_group": "adult",
  "crowd_preference": "quiet",
  "season": "winter"
}
```

### 8. Database Schema

Trips use a 4-level hierarchy:

```
trips
  └── trip_destinations (Cairo 3 days, Luxor 2 days, …)
        └── trip_days (Day 1, Day 2, …)
              └── trip_activities (Pyramids 4h, Museum 3h, …)
```

## Project Structure

```
backend/
├── cmd/
│   ├── server/main.go           # App entry point
│   └── seed/main.go             # Database seeder
├── internal/
│   ├── config/                  # Environment config
│   ├── database/                # SQLite connection + migration
│   ├── handlers/                # HTTP route handlers
│   │   ├── auth_handler.go      # Auth endpoints
│   │   ├── chat_handler.go      # Ask WAYNX chat endpoints
│   │   ├── place_handler.go     # Places + favorites
│   │   ├── trip_handler.go      # Trip planning
│   │   └── user_handler.go      # User profile + stats
│   ├── middleware/              # Auth, error recovery
│   ├── models/                  # GORM database models
│   │   ├── chat.go              # ChatSession, ChatMessage, StringSlice
│   │   ├── notification.go      # Notification (model only, not yet wired)
│   │   ├── place.go
│   │   ├── saved_place.go       # SavedPlace (favorites)
│   │   ├── trip.go              # Trip, TripDestination, TripDay, TripActivity
│   │   └── user.go              # User, Preferences
│   ├── repository/              # Database queries
│   │   ├── chat_repo.go
│   │   ├── notification_repo.go
│   │   ├── place_repo.go
│   │   ├── saved_place_repo.go
│   │   ├── trip_repo.go
│   │   └── user_repo.go
│   ├── services/                # Business logic + AI client
│   │   ├── ai_client.go         # Gemini AI client (recommend + ask)
│   │   ├── auth_service.go
│   │   ├── chat_service.go      # Ask WAYNX logic
│   │   ├── notification_service.go
│   │   ├── place_service.go
│   │   ├── saved_place_service.go
│   │   ├── trip_service.go
│   │   └── user_service.go
│   └── utils/                   # JWT, response helpers
├── ai demo/
│   └── waynx_demo.html          # Browser-based AI recommendation simulator
├── .env.example
├── go.mod
└── go.sum
```
