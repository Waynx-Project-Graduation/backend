package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// sanitizeJSON aggressively strips markdown formatting from AI responses
func sanitizeJSON(input string) string {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "```json") {
		input = strings.TrimPrefix(input, "```json")
	} else if strings.HasPrefix(input, "```") {
		input = strings.TrimPrefix(input, "```")
	}
	input = strings.TrimSuffix(input, "```")
	return strings.TrimSpace(input)
}

// AIClient communicates with the WAYNX recommendation API and Google Gemini
type AIClient struct {
	geminiAPIKey   string
	waynxBaseURL   string // e.g. "https://waynx-api-production.up.railway.app"
	timeoutSeconds int
	httpClient     *http.Client
}

func NewAIClient(geminiAPIKey, waynxBaseURL string, timeoutSeconds int) *AIClient {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	return &AIClient{
		geminiAPIKey:   geminiAPIKey,
		waynxBaseURL:   strings.TrimRight(waynxBaseURL, "/"),
		timeoutSeconds: timeoutSeconds,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// ─── Trip Recommendation Types ───────────────────────────────────────────────

// RecommendRequest is the internal request sent to the WAYNX AI service
type RecommendRequest struct {
	Interests        []string `json:"interests"`
	TravelCompanion  string   `json:"travel_companion"`
	Budget           string   `json:"budget"`
	AgeGroup         string   `json:"age_group"`
	CrowdPreference  string   `json:"crowd_preference"`
	Season           string   `json:"season"`
	TripDurationDays int      `json:"trip_duration_days"`
}

// ── WAYNX /itinerary API response types ─────────────────────────────────────

// waynxItineraryRequest is the body sent to POST /itinerary
type waynxItineraryRequest struct {
	User RecommendRequest `json:"user"`
}

// waynxItineraryResponse is the full JSON returned by POST /itinerary
type waynxItineraryResponse struct {
	Status    string              `json:"status"`
	TotalDays int                 `json:"total_days"`
	Itinerary []waynxDestination  `json:"itinerary"`
	Summary   waynxSummary        `json:"summary"`
}

type waynxDestination struct {
	DestinationIndex        int             `json:"destination_index"`
	City                    string          `json:"city"`
	CategoryFocus           string          `json:"category_focus"`
	DaysAllocated           int             `json:"days_allocated"`
	TravelHoursFromPrevious float64         `json:"travel_hours_from_previous"`
	DayPlan                 []waynxDayPlan  `json:"day_plan"`
}

type waynxDayPlan struct {
	Day        int              `json:"day"`
	Activities []waynxActivity  `json:"activities"`
	HoursUsed  int              `json:"hours_used"`
	FreeHours  int              `json:"free_hours"`
}

type waynxActivity struct {
	PlaceID     uint    `json:"place_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	Hours       int     `json:"hours"`
	Description string  `json:"description"`
	MatchPct    float64 `json:"match_pct"`
}

type waynxSummary struct {
	TotalDays       int      `json:"total_days"`
	DaysPlanned     int      `json:"days_planned"`
	Destinations    int      `json:"destinations"`
	TotalActivities int      `json:"total_activities"`
	Cities          []string `json:"cities"`
}

// ── Normalized internal types (used by trip_service.go) ─────────────────────

// RecommendResponse is the normalized response consumed by trip_service.go
type RecommendResponse struct {
	Recommendations []AIRecommendation `json:"recommendations"`
	Plan            AIPlan             `json:"plan"`
}

type AIRecommendation struct {
	PlaceID        uint    `json:"place_id"`
	Name           string  `json:"name"`
	City           string  `json:"city"`
	Category       string  `json:"category"`
	BudgetLevel    string  `json:"budget_level"`
	Rating         float64 `json:"rating"`
	DurationNeeded int     `json:"duration_needed"`
	Description    string  `json:"description"`
	MatchScore     int     `json:"match_score"` // 0–100
}

type AIPlan struct {
	TotalDays    int             `json:"total_days"`
	Destinations []AIDestination `json:"destinations"`
}

type AIDestination struct {
	City          string      `json:"city"`
	Days          int         `json:"days"`
	Category      string      `json:"category"`
	DailySchedule []AIPlanDay `json:"daily_schedule"`
}

type AIPlanDay struct {
	DayNumber  int              `json:"day_number"`
	HoursUsed  int              `json:"hours_used"`
	Activities []AIPlanActivity `json:"activities"`
}

type AIPlanActivity struct {
	PlaceID       uint    `json:"place_id"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	DurationHours int     `json:"duration_hours"`
	Rating        float64 `json:"rating"`
}

// ─── Chat / Ask WAYNX Types ─────────────────────────────────────────────────

type ChatRequest struct {
	Message string        `json:"message"`
	History []ChatHistory `json:"history,omitempty"`
}

type ChatHistory struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

type ChatResponse struct {
	Answer         string   `json:"answer"`
	IsVerified     bool     `json:"is_verified"`
	RelatedPlaces  []string `json:"related_places,omitempty"`
	SuggestedTitle string   `json:"suggested_title,omitempty"`
}

// ─── Trip Recommendation (calls WAYNX /itinerary) ───────────────────────────

// GetRecommendation calls the external WAYNX API /itinerary endpoint and returns
// the plan normalized into the internal RecommendResponse format.
func (c *AIClient) GetRecommendation(req RecommendRequest) (*RecommendResponse, error) {
	if c.waynxBaseURL == "" {
		return nil, fmt.Errorf("WAYNX API URL is not configured")
	}

	// Build the request body that the WAYNX API expects: { "user": { ... } }
	body := waynxItineraryRequest{User: req}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WAYNX request: %w", err)
	}

	url := c.waynxBaseURL + "/itinerary"
	log.Printf("Calling WAYNX API: POST %s (duration=%d days)", url, req.TripDurationDays)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutSeconds)*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create WAYNX request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("WAYNX API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAYNX response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("WAYNX API returned status %d: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("WAYNX API returned status %d", resp.StatusCode)
	}

	// Parse the WAYNX-specific response
	var waynxResp waynxItineraryResponse
	if err := json.Unmarshal(respBody, &waynxResp); err != nil {
		log.Printf("Failed to parse WAYNX response: %s", string(respBody))
		return nil, fmt.Errorf("failed to parse WAYNX response: %w", err)
	}

	if waynxResp.Status != "ok" {
		return nil, fmt.Errorf("WAYNX API returned status: %s", waynxResp.Status)
	}

	// Convert the WAYNX response → internal RecommendResponse
	result := c.convertWAYNXResponse(&waynxResp)

	log.Printf("WAYNX API returned %d destinations, %d total activities",
		len(result.Plan.Destinations), waynxResp.Summary.TotalActivities)

	return result, nil
}

// convertWAYNXResponse transforms the WAYNX /itinerary response into the
// internal RecommendResponse format that trip_service.go already understands.
func (c *AIClient) convertWAYNXResponse(waynx *waynxItineraryResponse) *RecommendResponse {
	result := &RecommendResponse{
		Plan: AIPlan{
			TotalDays:    waynx.TotalDays,
			Destinations: make([]AIDestination, 0, len(waynx.Itinerary)),
		},
		Recommendations: make([]AIRecommendation, 0),
	}

	seen := make(map[uint]bool) // track unique places for recommendations list

	for _, dest := range waynx.Itinerary {
		aiDest := AIDestination{
			City:          dest.City,
			Days:          dest.DaysAllocated,
			Category:      dest.CategoryFocus,
			DailySchedule: make([]AIPlanDay, 0, len(dest.DayPlan)),
		}

		for _, day := range dest.DayPlan {
			aiDay := AIPlanDay{
				DayNumber:  day.Day,
				HoursUsed:  day.HoursUsed,
				Activities: make([]AIPlanActivity, 0, len(day.Activities)),
			}

			for _, act := range day.Activities {
				aiDay.Activities = append(aiDay.Activities, AIPlanActivity{
					PlaceID:       act.PlaceID,
					Name:          act.Name,
					Category:      act.Category,
					DurationHours: act.Hours,
					Rating:        act.Rating,
				})

				// Also collect unique places into the recommendations list
				if !seen[act.PlaceID] {
					seen[act.PlaceID] = true
					result.Recommendations = append(result.Recommendations, AIRecommendation{
						PlaceID:        act.PlaceID,
						Name:           act.Name,
						City:           dest.City,
						Category:       act.Category,
						Rating:         act.Rating,
						DurationNeeded: act.Hours,
						Description:    act.Description,
						MatchScore:     int(act.MatchPct),
					})
				}
			}

			aiDest.DailySchedule = append(aiDest.DailySchedule, aiDay)
		}

		result.Plan.Destinations = append(result.Plan.Destinations, aiDest)
	}

	return result
}

// ─── Chat / Ask WAYNX (still uses Gemini) ───────────────────────────────────

// AskQuestion sends a chat message to Gemini and gets a conversational response
func (c *AIClient) AskQuestion(req ChatRequest) (*ChatResponse, error) {
	if c.geminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutSeconds)*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(c.geminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	cs := model.StartChat()

	// Pre-fill history
	for _, msg := range req.History {
		if msg.Role == "user" || msg.Role == "model" || msg.Role == "assistant" {
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}
			cs.History = append(cs.History, &genai.Content{
				Role:  role,
				Parts: []genai.Part{genai.Text(msg.Content)},
			})
		}
	}

	prompt := fmt.Sprintf(`You are WAYNX, an expert AI travel planner for Egypt.
The user is asking: "%s"

Reply ONLY with a valid JSON object matching this structure:
{
  "answer": (string, your detailed response to the user),
  "is_verified": (boolean, true if the information is factually verified),
  "related_places": (array of strings, optional place IDs related to the answer, e.g. ["1", "5"]),
  "suggested_title": (string, a short 3-4 word title summarizing this entire conversation, optional)
}`, req.Message)

	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to send message to gemini: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned an empty response")
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("gemini response part is not text")
	}

	var result ChatResponse
	cleanJSON := sanitizeJSON(string(text))
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		log.Printf("Failed to unmarshal JSON from Gemini: %s", cleanJSON)
		return nil, fmt.Errorf("failed to parse AI JSON response: %w", err)
	}

	return &result, nil
}
