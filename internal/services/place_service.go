package services

import (
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type PlaceService struct {
	placeRepo *repository.PlaceRepository
	aiClient  *AIClient
}

func NewPlaceService(placeRepo *repository.PlaceRepository, aiClient *AIClient) *PlaceService {
	return &PlaceService{
		placeRepo: placeRepo,
		aiClient:  aiClient,
	}
}

func (s *PlaceService) GetPlace(id uint) (*models.Place, error) {
	return s.placeRepo.FindByID(id)
}

// ListPlacesAdvanced supports the full UI filter set
func (s *PlaceService) ListPlacesAdvanced(filter repository.PlaceFilter) ([]models.Place, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 500 {
		filter.PerPage = 10
	}
	return s.placeRepo.ListWithFilters(filter)
}

func (s *PlaceService) PopularPlaces(limit int) ([]models.Place, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.placeRepo.Popular(limit)
}

func (s *PlaceService) SearchPlaces(q string, limit int) ([]models.Place, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.placeRepo.Search(q, limit)
}

// ListCategories returns all available place categories
func (s *PlaceService) ListCategories() ([]repository.CategoryInfo, error) {
	return s.placeRepo.ListCategories()
}

// TrendingSearches returns popular trending search terms
func (s *PlaceService) TrendingSearches(limit int) ([]string, error) {
	if limit < 1 || limit > 20 {
		limit = 8
	}
	return s.placeRepo.TrendingSearches(limit)
}

func (s *PlaceService) UpdateThumbnail(id uint, url string) error {
	return s.placeRepo.UpdateThumbnail(id, url)
}

func (s *PlaceService) ListCities() ([]repository.CityInfo, error) {
	return s.placeRepo.ListCities()
}

func (s *PlaceService) RecommendPlaces(req RecommendRequest) ([]AIRecommendation, error) {
	resp, err := s.aiClient.GetRecommendation(req)
	if err != nil {
		return nil, err
	}
	return resp.Recommendations, nil
}
