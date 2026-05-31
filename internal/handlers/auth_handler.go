package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Register(input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, resp)
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Login(input)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, resp)
}

// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "refresh_token is required")
		return
	}

	tokens, err := h.authService.RefreshToken(body.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, tokens)
}

// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// POST /api/auth/google
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	var body struct {
		Email      string `json:"email" binding:"required,email"`
		FullName   string `json:"full_name" binding:"required"`
		ProviderID string `json:"provider_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.GoogleAuth(body.Email, body.FullName, body.ProviderID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, resp)
}

// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input services.ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.ForgotPassword(input)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, resp)
}

// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input services.ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ResetPassword(input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "password reset successfully"})
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is handled client-side by discarding tokens.
	// This endpoint exists for API completeness and can be extended
	// to support token blacklisting if needed.
	utils.Success(c, gin.H{"message": "logged out successfully"})
}
