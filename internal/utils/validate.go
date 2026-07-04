package utils

import "strings"

// Allowed enum values for user preference fields. Kept here so both the user
// and trip layers can validate against a single source of truth.
var (
	ValidBudgets          = []string{"low", "medium", "high"}
	ValidAgeGroups        = []string{"kid", "teen", "adult", "senior"}
	ValidSeasons          = []string{"winter", "spring", "summer", "autumn", "any"}
	ValidCrowdPreferences = []string{"quiet", "moderate", "crowded", "no_preference"}
	ValidCompanions       = []string{"solo", "couple", "family", "friends"}
	ValidInterests        = []string{
		"history", "adventure", "nature", "culture", "food",
		"relaxation", "nightlife", "shopping", "beach", "religious",
	}
)

// IsOneOf reports whether value (case-insensitive) is in the allowed set.
// An empty value is considered valid so callers can treat it as "not provided".
func IsOneOf(value string, allowed []string) bool {
	if value == "" {
		return true
	}
	v := strings.ToLower(strings.TrimSpace(value))
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// AllOneOf reports whether every value in values is in the allowed set.
func AllOneOf(values []string, allowed []string) bool {
	for _, v := range values {
		if !IsOneOf(v, allowed) {
			return false
		}
	}
	return true
}
