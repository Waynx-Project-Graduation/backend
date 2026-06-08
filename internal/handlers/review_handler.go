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

// POST /api/places/:id/reviews
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

// GET /api/places/:id/reviews
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

// PUT /api/places/:id/reviews/:reviewId
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

// DELETE /api/places/:id/reviews/:reviewId
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
