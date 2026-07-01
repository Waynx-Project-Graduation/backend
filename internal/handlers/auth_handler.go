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

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body services.RegisterInput true "Registration info"
// @Success      201  {object}  services.AuthResponse
// @Failure      400  {object}  utils.Response
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Login user
// @Description  Authenticates a user and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body services.LoginInput true "Login info"
// @Success      200  {object}  services.AuthResponse
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Router       /auth/login [post]
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

// RefreshToken godoc
// @Summary      Refresh tokens
// @Description  Exchanges a refresh token for a new pair of access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{refresh_token=string} true "Refresh Token"
// @Success      200  {object}  utils.TokenPair
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Router       /auth/refresh [post]
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

// Me godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the currently authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  services.UserResponse
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, user)
}

// GoogleAuth godoc
// @Summary      Google authentication
// @Description  Authenticates a user via Google OAuth
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{email=string,full_name=string,provider_id=string} true "Google Auth Info"
// @Success      200  {object}  services.AuthResponse
// @Failure      400  {object}  utils.Response
// @Router       /auth/google [post]
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

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Initiates password reset process
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body services.ForgotPasswordInput true "Email for password reset"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets user password using a token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body services.ResetPasswordInput true "Reset password info"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /auth/reset-password [post]
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

// Logout godoc
// @Summary      Logout user
// @Description  Logs out the current user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is handled client-side by discarding tokens.
	// This endpoint exists for API completeness and can be extended
	// to support token blacklisting if needed.
	utils.Success(c, gin.H{"message": "logged out successfully"})
}
