package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type ReviewHandler struct {
	reviewService *services.ReviewService
}

func NewReviewHandler(reviewService *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// CreateReview godoc
// @Summary      Create review
// @Description  Create a new review for a place
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Place ID"
// @Param        input body services.CreateReviewInput true "Review data"
// @Success      201  {object}  models.PlaceReview
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /places/{id}/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := getUserID(c)
	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	var input services.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	review, err := h.reviewService.CreateReview(userID, uint(placeID), input)
	if err != nil {
		if err.Error() == "place not found" {
			utils.NotFound(c, err.Error())
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, review)
}

// ListReviews godoc
// @Summary      List reviews
// @Description  Get a paginated list of reviews for a place
// @Tags         reviews
// @Produce      json
// @Param        id path int true "Place ID"
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(10)
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /places/{id}/reviews [get]
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	placeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid place ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	reviews, total, err := h.reviewService.ListReviews(uint(placeID), page, perPage)
	if err != nil {
		utils.InternalError(c, "failed to list reviews")
		return
	}

	utils.SuccessWithMeta(c, reviews, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// UpdateReview godoc
// @Summary      Update review
// @Description  Update an existing review
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Place ID"
// @Param        reviewId path string true "Review ID (UUID)"
// @Param        input body services.UpdateReviewInput true "Review data"
// @Success      200  {object}  models.PlaceReview
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /places/{id}/reviews/{reviewId} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID := getUserID(c)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		utils.BadRequest(c, "invalid review ID")
		return
	}

	var input services.UpdateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	review, err := h.reviewService.UpdateReview(reviewID, userID, input)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, review)
}

// DeleteReview godoc
// @Summary      Delete review
// @Description  Delete an existing review
// @Tags         reviews
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Place ID"
// @Param        reviewId path string true "Review ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /places/{id}/reviews/{reviewId} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID := getUserID(c)
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		utils.BadRequest(c, "invalid review ID")
		return
	}

	if err := h.reviewService.DeleteReview(reviewID, userID); err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "review deleted successfully"})
}
