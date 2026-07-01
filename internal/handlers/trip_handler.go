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

// CreateTrip godoc
// @Summary      Create a new trip
// @Description  Generates an AI-powered trip itinerary based on user input
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body services.CreateTripInput true "Trip creation parameters"
// @Success      201  {object}  models.Trip
// @Failure      400  {object}  utils.Response
// @Router       /trips [post]
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

// ListTrips godoc
// @Summary      List user trips
// @Description  Get a list of trips for the authenticated user
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Trip
// @Failure      401  {object}  utils.Response
// @Router       /trips [get]
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

// GetTrip godoc
// @Summary      Get trip details
// @Description  Get full details of a specific trip, including days and activities
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Trip ID"
// @Success      200  {object}  models.Trip
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id} [get]
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

// UpdateTrip godoc
// @Summary      Update trip
// @Description  Update details of an existing trip
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        input body services.UpdateTripInput true "Trip update info"
// @Success      200  {object}  models.Trip
// @Failure      400  {object}  utils.Response
// @Router       /trips/{id} [put]
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

// DeleteTrip godoc
// @Summary      Delete trip
// @Description  Delete a trip and its associated data
// @Tags         trips
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /trips/{id} [delete]
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

// RegenerateItinerary godoc
// @Summary      Regenerate itinerary
// @Description  Use AI to regenerate the trip itinerary
// @Tags         trips
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Success      200  {object}  models.Trip
// @Failure      400  {object}  utils.Response
// @Failure      503  {object}  utils.Response
// @Router       /trips/{id}/regenerate [post]
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

// CreateActivity godoc
// @Summary      Create activity
// @Description  Add a new activity to a trip
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        input body services.CreateActivityInput true "Activity info"
// @Success      201  {object}  models.TripActivity
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Router       /trips/{id}/activities [post]
func (h *TripHandler) CreateActivity(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	var input services.CreateActivityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	activity, err := h.tripService.CreateActivity(tripID, userID, input)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, activity)
}

// UpdateActivity godoc
// @Summary      Update activity
// @Description  Update an existing activity
// @Tags         trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        activityId path string true "Activity ID (UUID)"
// @Param        input body services.UpdateActivityInput true "Activity update info"
// @Success      200  {object}  models.TripActivity
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/activities/{activityId} [put]
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

// DeleteActivity godoc
// @Summary      Delete activity
// @Description  Delete an activity from a trip
// @Tags         trips
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        activityId path string true "Activity ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/activities/{activityId} [delete]
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
