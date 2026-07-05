package handlers

import (
	"errors"
	"log"
	"strconv"
	"strings"

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

// mapUserError translates service sentinel errors into the correct HTTP status.
func mapUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		utils.NotFound(c, "user not found")
	case errors.Is(err, services.ErrIncorrectPassword):
		utils.Unauthorized(c, "incorrect current password")
	case errors.Is(err, services.ErrPasswordNotAllowed):
		utils.BadRequest(c, err.Error())
	default:
		utils.BadRequest(c, err.Error())
	}
}

// GET /api/users/profile — Get the authenticated user's full profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		mapUserError(c, err)
		return
	}

	utils.Success(c, user)
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
		mapUserError(c, err)
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
		mapUserError(c, err)
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
		mapUserError(c, err)
		return
	}

	utils.Success(c, gin.H{"message": "password updated successfully"})
}

// GET /api/users/stats — Get user profile statistics
func (h *UserHandler) GetStats(c *gin.Context) {
	userID := getUserID(c)

	stats, err := h.userService.GetStats(userID)
	if err != nil {
		mapUserError(c, err)
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

// PUT /api/users/avatar — Update user avatar URL via file upload
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

	if !isAllowedImage(header.Filename, header.Header.Get("Content-Type")) {
		utils.BadRequest(c, "invalid file type: only JPEG, PNG, WebP, or GIF images are allowed")
		return
	}

	// Upload to Cloudinary
	url, err := h.cloudinaryService.UploadImage(c.Request.Context(), file, "trip-planner/avatars")
	if err != nil {
		log.Printf("Failed to upload avatar to Cloudinary: %v", err)
		utils.InternalError(c, "failed to upload image to Cloudinary")
		return
	}

	profileInput := services.UpdateProfileInput{
		AvatarURL: &url,
	}

	user, err := h.userService.UpdateProfile(userID, profileInput)
	if err != nil {
		mapUserError(c, err)
		return
	}

	utils.Success(c, user)
}

// isAllowedImage validates an uploaded file is an image by both its declared
// content-type and its file extension.
func isAllowedImage(filename, contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	}

	ct := strings.ToLower(strings.TrimSpace(contentType))
	if !allowedTypes[ct] {
		return false
	}

	name := strings.ToLower(filename)
	dot := strings.LastIndex(name, ".")
	if dot < 0 {
		return false
	}
	return allowedExts[name[dot:]]
}

// DELETE /api/users/account — Delete user account (soft delete)
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
