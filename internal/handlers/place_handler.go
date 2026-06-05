package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type PlaceHandler struct {
	placeService      *services.PlaceService
	savedPlaceService *services.SavedPlaceService
	cloudinaryService *services.CloudinaryService
}

func NewPlaceHandler(placeService *services.PlaceService, savedPlaceService *services.SavedPlaceService, cloudinaryService *services.CloudinaryService) *PlaceHandler {
	return &PlaceHandler{
		placeService:      placeService,
		savedPlaceService: savedPlaceService,
		cloudinaryService: cloudinaryService,
	}
}

// GET /api/places — List places with advanced filtering
func (h *PlaceHandler) ListPlaces(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	// Check if advanced filters are used
	citiesParam := c.QueryArray("cities[]")
	city := c.Query("city") // backward-compatible single city
	category := c.Query("category")
	sortBy := c.Query("sort_by")
	search := c.Query("search")

	// Parse budget levels: "low,medium" or budget_level[]=low&budget_level[]=medium
	var budgetLevels []string
	budgetLevelParams := c.QueryArray("budget_level[]")
	if len(budgetLevelParams) == 0 {
		if bl := c.Query("budget_level"); bl != "" {
			for _, b := range strings.Split(bl, ",") {
				budgetLevels = append(budgetLevels, strings.TrimSpace(b))
			}
		}
	} else {
		budgetLevels = budgetLevelParams
	}

	bestSeason := c.Query("best_season")
	crowdLevel := c.Query("crowd_level")
	suitableFor := c.Query("suitable_for")
	suitableAge := c.Query("suitable_age")

	// Merge single city into cities list for backward compatibility
	if city != "" && len(citiesParam) == 0 {
		citiesParam = []string{city}
	}

	filter := repository.PlaceFilter{
		Cities:      citiesParam,
		Category:    category,
		BudgetLevel: budgetLevels,
		BestSeason:  bestSeason,
		CrowdLevel:  crowdLevel,
		SuitableFor: suitableFor,
		SuitableAge: suitableAge,
		SortBy:      sortBy,
		Search:      search,
		Page:        page,
		PerPage:     perPage,
	}

	places, total, err := h.placeService.ListPlacesAdvanced(filter)
	if err != nil {
		utils.InternalError(c, "failed to list places")
		return
	}

	utils.SuccessWithMeta(c, places, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GET /api/places/:id — Get place details
func (h *PlaceHandler) GetPlace(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	place, err := h.placeService.GetPlace(uint(id))
	if err != nil {
		utils.NotFound(c, "place not found")
		return
	}

	utils.Success(c, place)
}

// GET /api/places/popular — Popular places
func (h *PlaceHandler) PopularPlaces(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	places, err := h.placeService.PopularPlaces(limit)
	if err != nil {
		utils.InternalError(c, "failed to get popular places")
		return
	}

	utils.Success(c, places)
}

// GET /api/places/search — Search places
func (h *PlaceHandler) SearchPlaces(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		utils.BadRequest(c, "search query 'q' is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	places, err := h.placeService.SearchPlaces(q, limit)
	if err != nil {
		utils.InternalError(c, "failed to search places")
		return
	}

	utils.Success(c, places)
}

// GET /api/places/categories — List all available place categories
func (h *PlaceHandler) ListCategories(c *gin.Context) {
	categories, err := h.placeService.ListCategories()
	if err != nil {
		utils.InternalError(c, "failed to list categories")
		return
	}

	utils.Success(c, categories)
}

// GET /api/places/trending — Get trending search terms
func (h *PlaceHandler) TrendingSearches(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))

	trending, err := h.placeService.TrendingSearches(limit)
	if err != nil {
		utils.InternalError(c, "failed to get trending searches")
		return
	}

	utils.Success(c, trending)
}

// POST /api/places/:id/save — Save a place to favorites
func (h *PlaceHandler) SavePlace(c *gin.Context) {
	userID := getUserID(c)
	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	if err := h.savedPlaceService.SavePlace(userID, uint(placeID)); err != nil {
		if err.Error() == "place not found" {
			utils.NotFound(c, err.Error())
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, gin.H{"message": "place saved successfully"})
}

// DELETE /api/places/:id/save — Remove a place from favorites
func (h *PlaceHandler) UnsavePlace(c *gin.Context) {
	userID := getUserID(c)
	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	if err := h.savedPlaceService.UnsavePlace(userID, uint(placeID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "place unsaved successfully"})
}

// POST /api/places/:id/photo — Upload a place photo
func (h *PlaceHandler) UploadPlacePhoto(c *gin.Context) {
	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		utils.BadRequest(c, "photo file is required")
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 { // 5MB limit
		utils.BadRequest(c, "file size exceeds 5MB limit")
		return
	}

	// Upload to Cloudinary
	url, err := h.cloudinaryService.UploadImage(c.Request.Context(), file, "trip-planner/places")
	if err != nil {
		utils.InternalError(c, "failed to upload image to Cloudinary")
		return
	}

	// Update Database
	if err := h.placeService.UpdateThumbnail(uint(placeID), url); err != nil {
		utils.InternalError(c, "failed to update place photo in database")
		return
	}

	utils.Success(c, gin.H{
		"message":       "place photo uploaded successfully",
		"thumbnail_url": url,
	})
}
