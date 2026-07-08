package repository

import (
	"strings"

	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type PlaceRepository struct {
	db *gorm.DB
}

func NewPlaceRepository(db *gorm.DB) *PlaceRepository {
	return &PlaceRepository{db: db}
}

func (r *PlaceRepository) Create(place *models.Place) error {
	return r.db.Create(place).Error
}

func (r *PlaceRepository) FindByID(id uint) (*models.Place, error) {
	var place models.Place
	err := r.db.Where("id = ?", id).First(&place).Error
	if err != nil {
		return nil, err
	}
	return &place, nil
}

// PlaceFilter holds all the advanced filter parameters the UI supports
type PlaceFilter struct {
	Cities      []string // multi-city filter (Cairo, Luxor, etc.)
	Category    string   // category filter (history, adventure, etc.)
	BudgetLevel []string // budget: low, medium, high
	BestSeason  string   // winter, summer, spring, autumn, any
	CrowdLevel  string   // quiet, moderate, crowded
	SuitableFor string   // family, couple, solo, friends
	SuitableAge string   // kid, teen, adult, senior
	SortBy      string   // "rating", "name", "duration_asc", "duration_desc"
	Search      string   // search query
	Page        int
	PerPage     int
}

// ListWithFilters returns places matching the advanced filter criteria
func (r *PlaceRepository) ListWithFilters(filter PlaceFilter) ([]models.Place, int64, error) {
	var places []models.Place
	var total int64

	query := r.db.Model(&models.Place{})

	// Multi-city filter
	if len(filter.Cities) > 0 {
		query = query.Where("LOWER(city) IN (?)", toLowerCase(filter.Cities))
	}

	// Category filter
	if filter.Category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", filter.Category)
	}

	// Budget level filter
	if len(filter.BudgetLevel) > 0 {
		query = query.Where("LOWER(budget_level) IN (?)", toLowerCase(filter.BudgetLevel))
	}

	// Best season filter
	if filter.BestSeason != "" {
		query = query.Where("LOWER(best_season) = LOWER(?)", filter.BestSeason)
	}

	// Crowd level filter
	if filter.CrowdLevel != "" {
		query = query.Where("LOWER(crowd_level) = LOWER(?)", filter.CrowdLevel)
	}

	// Suitable for filter (searches within the comma-separated field)
	if filter.SuitableFor != "" {
		query = query.Where("LOWER(suitable_for) LIKE LOWER(?)", "%"+filter.SuitableFor+"%")
	}

	// Suitable age filter (searches within the comma-separated field)
	if filter.SuitableAge != "" {
		query = query.Where("LOWER(suitable_age) LIKE LOWER(?)", "%"+filter.SuitableAge+"%")
	}

	// Search filter
	if filter.Search != "" {
		query = query.Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?) OR LOWER(category) LIKE LOWER(?) OR LOWER(city) LIKE LOWER(?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%",
		)
	}

	query.Count(&total)

	// Sorting
	switch filter.SortBy {
	case "duration_asc":
		query = query.Order("duration_needed ASC")
	case "duration_desc":
		query = query.Order("duration_needed DESC")
	case "name":
		query = query.Order("name ASC")
	default: // "rating" or default
		query = query.Order("rating DESC")
	}

	// Pagination
	offset := (filter.Page - 1) * filter.PerPage
	err := query.Offset(offset).Limit(filter.PerPage).Find(&places).Error
	return places, total, err
}

// List is the legacy list method kept for backward compatibility
func (r *PlaceRepository) List(city, category string, page, perPage int) ([]models.Place, int64, error) {
	filter := PlaceFilter{
		Category: category,
		Page:     page,
		PerPage:  perPage,
	}
	if city != "" {
		filter.Cities = []string{city}
	}
	return r.ListWithFilters(filter)
}

func (r *PlaceRepository) Popular(limit int) ([]models.Place, error) {
	var places []models.Place
	err := r.db.Order("rating DESC").Limit(limit).Find(&places).Error
	return places, err
}

func (r *PlaceRepository) Search(q string, limit int) ([]models.Place, error) {
	var places []models.Place
	err := r.db.Where(
		"LOWER(name) LIKE LOWER(?) OR LOWER(city) LIKE LOWER(?) OR LOWER(category) LIKE LOWER(?)",
		"%"+q+"%", "%"+q+"%", "%"+q+"%").
		Limit(limit).Order("rating DESC").Find(&places).Error
	return places, err
}

// ListCategories returns all distinct categories with their counts and an image url
func (r *PlaceRepository) ListCategories() ([]CategoryInfo, error) {
	var results []CategoryInfo
	err := r.db.Model(&models.Place{}).
		Select("category, COUNT(*) as count, MIN(thumbnail_url) as image_url").
		Where("category != ''").
		Group("category").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// TrendingSearches returns popular search terms based on high-rated places
func (r *PlaceRepository) TrendingSearches(limit int) ([]string, error) {
	var names []string
	err := r.db.Model(&models.Place{}).
		Select("name").
		Order("rating DESC").
		Limit(limit).
		Pluck("name", &names).Error
	return names, err
}

// CategoryInfo represents a category with its count and an image url
type CategoryInfo struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
	ImageURL string `json:"image_url"`
}

// CityInfo represents a city with its count of places
type CityInfo struct {
	City  string `json:"city"`
	Count int64  `json:"count"`
}

// ListCities returns all distinct cities with their place counts
func (r *PlaceRepository) ListCities() ([]CityInfo, error) {
	var results []CityInfo
	err := r.db.Model(&models.Place{}).
		Select("city, COUNT(*) as count").
		Where("city != ''").
		Group("city").
		Order("count DESC").
		Find(&results).Error
	return results, err
}

// toLowerCase converts a string slice to lowercase
func toLowerCase(items []string) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = strings.ToLower(item)
	}
	return result
}
