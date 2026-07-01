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
	cloudinaryService *services.CloudinaryService
}

func NewUserHandler(userService *services.UserService, authService *services.AuthService, savedPlaceService *services.SavedPlaceService, cloudinaryService *services.CloudinaryService) *UserHandler {
	return &UserHandler{
		userService:       userService,
		authService:       authService,
		savedPlaceService: savedPlaceService,
		cloudinaryService: cloudinaryService,
	}
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body services.UpdateProfileInput true "Profile info"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/profile [put]
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

// UpdatePreferences godoc
// @Summary      Update travel preferences
// @Description  Updates the authenticated user's travel preferences
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body services.UpdatePreferencesInput true "Preferences info"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/preferences [put]
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

// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated user's password
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body services.ChangePasswordInput true "Password info"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/password [put]
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

// GetStats godoc
// @Summary      Get user stats
// @Description  Retrieves statistics for the authenticated user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/stats [get]
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := getUserID(c)

	stats, err := h.userService.GetStats(userID)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetSavedPlaces godoc
// @Summary      Get saved places
// @Description  Retrieves a paginated list of the user's saved places
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(10)
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/saved-places [get]
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

// UpdateAvatar godoc
// @Summary      Update avatar
// @Description  Uploads and updates the user's avatar image
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar formData file true "Avatar image file"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/avatar [put]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := getUserID(c)

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.BadRequest(c, "avatar file is required")
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 { // 5MB limit
		utils.BadRequest(c, "file size exceeds 5MB limit")
		return
	}

	// Upload to Cloudinary
	url, err := h.cloudinaryService.UploadImage(c.Request.Context(), file, "trip-planner/avatars")
	if err != nil {
		utils.InternalError(c, "failed to upload image to Cloudinary")
		return
	}

	profileInput := services.UpdateProfileInput{
		AvatarURL: url,
	}

	user, err := h.userService.UpdateProfile(userID, profileInput)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// DeleteAccount godoc
// @Summary      Delete account
// @Description  Soft deletes the authenticated user's account
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /users/account [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := getUserID(c)

	if err := h.userService.DeleteAccount(userID); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "account deleted successfully"})
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
