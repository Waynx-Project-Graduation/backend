package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type TripHandler struct {
	tripService *services.TripService
}

func NewTripHandler(tripService *services.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

// POST /api/trips
func (h *TripHandler) CreateTrip(c *gin.Context) {
	userID := getUserID(c)

	var input services.CreateTripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	trip, err := h.tripService.CreateTrip(userID, input)
	if err != nil {
		if err.Error() == "AI service unavailable" {
			utils.ServiceUnavailable(c, "AI service is currently unavailable. Trip saved as draft.")
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, trip)
}

// GET /api/trips
func (h *TripHandler) ListTrips(c *gin.Context) {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	trips, total, err := h.tripService.ListTrips(userID, page, perPage)
	if err != nil {
		utils.InternalError(c, "failed to list trips")
		return
	}

	utils.SuccessWithMeta(c, trips, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GET /api/trips/:id
func (h *TripHandler) GetTrip(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	trip, err := h.tripService.GetTrip(tripID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, "you don't have access to this trip")
			return
		}
		utils.NotFound(c, "trip not found")
		return
	}

	utils.Success(c, trip)
}

// PUT /api/trips/:id
func (h *TripHandler) UpdateTrip(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	var input services.UpdateTripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	trip, err := h.tripService.UpdateTrip(tripID, userID, input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, trip)
}

// DELETE /api/trips/:id
func (h *TripHandler) DeleteTrip(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	if err := h.tripService.DeleteTrip(tripID, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "trip deleted successfully"})
}

// POST /api/trips/:id/regenerate
func (h *TripHandler) RegenerateItinerary(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	trip, err := h.tripService.RegenerateItinerary(tripID, userID)
	if err != nil {
		utils.ServiceUnavailable(c, err.Error())
		return
	}

	utils.Success(c, trip)
}

// PUT /api/trips/:id/activities/:activityId
func (h *TripHandler) UpdateActivity(c *gin.Context) {
	userID := getUserID(c)
	activityID, err := uuid.Parse(c.Param("activityId"))
	if err != nil {
		utils.BadRequest(c, "invalid activity ID")
		return
	}

	var input services.UpdateActivityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	activity, err := h.tripService.UpdateActivity(activityID, userID, input)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, activity)
}

// DELETE /api/trips/:id/activities/:activityId
func (h *TripHandler) DeleteActivity(c *gin.Context) {
	userID := getUserID(c)
	activityID, err := uuid.Parse(c.Param("activityId"))
	if err != nil {
		utils.BadRequest(c, "invalid activity ID")
		return
	}

	if err := h.tripService.DeleteActivity(activityID, userID); err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "activity deleted successfully"})
}
