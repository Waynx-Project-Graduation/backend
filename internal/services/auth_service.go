package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// resetTokenEntry stores a password reset token mapping.
// In production, this would be stored in a database table.
type resetTokenEntry struct {
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
}

// In-memory reset token store (dev-mode only; use a DB table in production)
var (
	resetTokens   = make(map[string]resetTokenEntry) // token → entry
	resetTokensMu sync.Mutex
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *utils.JWTManager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *utils.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type RegisterInput struct {
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User   UserResponse     `json:"user"`
	Tokens *utils.TokenPair `json:"tokens"`
}

type UserResponse struct {
	ID             uuid.UUID          `json:"id"`
	FullName       string             `json:"full_name"`
	Email          string             `json:"email"`
	AuthProvider   string             `json:"auth_provider"`
	AvatarURL      string             `json:"avatar_url"`
	City           string             `json:"city"`
	ExplorerPoints int                `json:"explorer_points"`
	BadgeType      string             `json:"badge_type"`
	Preferences    models.Preferences `json:"preferences"`
	LastLogin      *time.Time         `json:"last_login,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
	// In production, token would be sent via email. For development, we return it.
	ResetToken string `json:"reset_token,omitempty"`
}

func toUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:             user.ID,
		FullName:       user.FullName,
		Email:          user.Email,
		AuthProvider:   user.AuthProvider,
		AvatarURL:      user.AvatarURL,
		City:           user.City,
		ExplorerPoints: user.ExplorerPoints,
		BadgeType:      user.BadgeType,
		Preferences:    user.Preferences,
		LastLogin:      user.LastLogin,
		CreatedAt:      user.CreatedAt,
	}
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
	// Check if email exists
	existing, _ := s.userRepo.FindByEmail(input.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		FullName:     input.FullName,
		Email:        input.Email,
		PasswordHash: string(hash),
		AuthProvider: "local",
		BadgeType:    "explorer",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	userResp := toUserResponse(user)
	return &AuthResponse{User: userResp, Tokens: tokens}, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, errors.New("failed to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	_ = s.userRepo.Update(user)

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	userResp := toUserResponse(user)
	return &AuthResponse{User: userResp, Tokens: tokens}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*utils.TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Verify user still exists
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return s.jwtManager.GenerateTokenPair(user.ID, user.Email)
}

func (s *AuthService) GetUserByID(id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *AuthService) GoogleAuth(email, fullName, providerID string) (*AuthResponse, error) {
	// Check if user exists with this provider ID
	user, err := s.userRepo.FindByProviderID("google", providerID)
	if err != nil {
		// User doesn't exist, create new
		user = &models.User{
			FullName:     fullName,
			Email:        email,
			AuthProvider: "google",
			ProviderID:   providerID,
			BadgeType:    "explorer",
		}
		if err := s.userRepo.Create(user); err != nil {
			// Maybe email already exists with local auth
			existing, _ := s.userRepo.FindByEmail(email)
			if existing != nil {
				return nil, errors.New("email already registered with different method")
			}
			return nil, errors.New("failed to create user")
		}
	}

	now := time.Now()
	user.LastLogin = &now
	_ = s.userRepo.Update(user)

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	userResp := toUserResponse(user)
	return &AuthResponse{User: userResp, Tokens: tokens}, nil
}

// ForgotPassword generates a password reset token and stores it.
// In production, the token would be sent via email; in dev mode, we return it.
func (s *AuthService) ForgotPassword(input ForgotPasswordInput) (*ForgotPasswordResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		// Don't reveal whether the email exists or not (security best practice)
		return &ForgotPasswordResponse{
			Message: "If an account with that email exists, a reset link has been sent.",
		}, nil
	}

	// Generate a secure reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errors.New("failed to generate reset token")
	}
	resetToken := hex.EncodeToString(tokenBytes)

	// Store the token with a 1-hour expiry
	resetTokensMu.Lock()
	resetTokens[resetToken] = resetTokenEntry{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	resetTokensMu.Unlock()

	return &ForgotPasswordResponse{
		Message:    "If an account with that email exists, a reset link has been sent.",
		ResetToken: resetToken, // Only returned in dev mode; remove in production
	}, nil
}

// ResetPassword resets the user's password using a valid reset token.
func (s *AuthService) ResetPassword(input ResetPasswordInput) error {
	if input.Token == "" {
		return errors.New("invalid reset token")
	}

	// Look up and validate the token
	resetTokensMu.Lock()
	entry, exists := resetTokens[input.Token]
	if exists {
		delete(resetTokens, input.Token) // single-use token
	}
	resetTokensMu.Unlock()

	if !exists {
		return errors.New("invalid or expired reset token")
	}
	if time.Now().After(entry.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the user's password
	return s.userRepo.UpdatePasswordHash(entry.UserID, string(hash))
}
