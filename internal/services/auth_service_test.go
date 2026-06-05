package services

import (
	"testing"
	"github.com/kemit/trip-planner/internal/repository"
)

func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtManager := setupJWTManager()
	authService := NewAuthService(userRepo, jwtManager)

	input := RegisterInput{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := authService.Register(input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", resp.User.Email)
	}
	if resp.Tokens == nil || resp.Tokens.AccessToken == "" {
		t.Errorf("Expected access token to be generated")
	}

	// Test duplicate email
	_, err = authService.Register(input)
	if err == nil {
		t.Errorf("Expected error for duplicate email registration")
	}
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtManager := setupJWTManager()
	authService := NewAuthService(userRepo, jwtManager)

	input := RegisterInput{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	authService.Register(input)

	// Valid login
	loginInput := LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	resp, err := authService.Login(loginInput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", resp.User.Email)
	}

	// Invalid password
	loginInput.Password = "wrongpassword"
	_, err = authService.Login(loginInput)
	if err == nil {
		t.Errorf("Expected error for invalid password")
	}

	// Invalid email
	loginInput.Email = "wrong@example.com"
	loginInput.Password = "password123"
	_, err = authService.Login(loginInput)
	if err == nil {
		t.Errorf("Expected error for invalid email")
	}
}

func TestAuthService_ForgotPasswordAndReset(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtManager := setupJWTManager()
	authService := NewAuthService(userRepo, jwtManager)

	input := RegisterInput{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	authService.Register(input)

	forgotResp, err := authService.ForgotPassword(ForgotPasswordInput{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("Expected no error for forgot password, got: %v", err)
	}
	if forgotResp.ResetToken == "" {
		t.Fatalf("Expected reset token to be generated")
	}

	// Reset password
	resetInput := ResetPasswordInput{
		Token:       forgotResp.ResetToken,
		NewPassword: "newpassword123",
	}
	err = authService.ResetPassword(resetInput)
	if err != nil {
		t.Fatalf("Expected no error resetting password, got: %v", err)
	}

	// Login with new password
	_, err = authService.Login(LoginInput{
		Email:    "test@example.com",
		Password: "newpassword123",
	})
	if err != nil {
		t.Fatalf("Expected successful login with new password, got: %v", err)
	}

	// Re-using token should fail
	err = authService.ResetPassword(resetInput)
	if err == nil {
		t.Fatalf("Expected error reusing reset token")
	}
}
