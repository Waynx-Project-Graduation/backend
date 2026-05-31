package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type UserHandler struct {
	userService       *services.UserService
	authService       *services.AuthService
	savedPlaceService *services.SavedPlaceService
}

func NewUserHandler(userService *services.UserService, authService *services.AuthService, savedPlaceService *services.SavedPlaceService) *UserHandler {
	return &UserHandler{
		userService:       userService,
		authService:       authService,
		savedPlaceService: savedPlaceService,
	}
}

// PUT /api/users/profile — Update user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserID(c)

	var input services.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, input)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// PUT /api/users/preferences — Update travel preferences
func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := getUserID(c)

	var input services.UpdatePreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdatePreferences(userID, input)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// PUT /api/users/password — Change password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := getUserID(c)

	var input services.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID, input); err != nil {
		if err.Error() == "incorrect current password" {
			utils.Unauthorized(c, err.Error())
			return
		}
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "password updated successfully"})
}

// GET /api/users/stats — Get user profile statistics
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := getUserID(c)

	stats, err := h.userService.GetStats(userID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GET /api/users/saved-places — List user's saved/bookmarked places
func (h *UserHandler) GetSavedPlaces(c *gin.Context) {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	saved, total, err := h.savedPlaceService.ListSavedPlaces(userID, page, perPage)
	if err != nil {
		utils.InternalError(c, "failed to list saved places")
		return
	}

	utils.SuccessWithMeta(c, saved, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// PUT /api/users/avatar — Update user avatar URL
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := getUserID(c)

	var input struct {
		AvatarURL string `json:"avatar_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	profileInput := services.UpdateProfileInput{
		AvatarURL: input.AvatarURL,
	}

	user, err := h.userService.UpdateProfile(userID, profileInput)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// getUserID extracts the user ID from gin context (set by auth middleware)
func getUserID(c *gin.Context) uuid.UUID {
	id, exists := c.Get("userID")
	if !exists {
		return uuid.Nil
	}
	uid, ok := id.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return uid
}
