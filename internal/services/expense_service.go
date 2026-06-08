package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type ExpenseService struct {
	expenseRepo *repository.ExpenseRepository
	tripRepo    *repository.TripRepository
}

func NewExpenseService(expenseRepo *repository.ExpenseRepository, tripRepo *repository.TripRepository) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		tripRepo:    tripRepo,
	}
}

type CreateExpenseInput struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type UpdateExpenseInput struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

func (s *ExpenseService) CreateExpense(userID, tripID uuid.UUID, input CreateExpenseInput) (*models.TripExpense, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}

	expense := &models.TripExpense{
		ID:          uuid.New(),
		TripID:      tripID,
		Amount:      input.Amount,
		Currency:    "EGP",
		Category:    input.Category,
		Description: input.Description,
		Date:        time.Now(),
	}

	if input.Currency != "" {
		expense.Currency = input.Currency
	}
	if input.Date != "" {
		if t, err := time.Parse("2006-01-02", input.Date); err == nil {
			expense.Date = t
		}
	}

	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, errors.New("failed to create expense")
	}
	return expense, nil
}

func (s *ExpenseService) ListExpenses(userID, tripID uuid.UUID, page, perPage int) ([]models.TripExpense, int64, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, 0, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, 0, errors.New("access denied")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return s.expenseRepo.ListByTripID(tripID, page, perPage)
}

func (s *ExpenseService) UpdateExpense(userID, tripID, expenseID uuid.UUID, input UpdateExpenseInput) (*models.TripExpense, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}

	expense, err := s.expenseRepo.FindByID(expenseID)
	if err != nil {
		return nil, errors.New("expense not found")
	}
	if expense.TripID != tripID {
		return nil, errors.New("expense does not belong to this trip")
	}

	if input.Amount > 0 {
		expense.Amount = input.Amount
	}
	if input.Currency != "" {
		expense.Currency = input.Currency
	}
	if input.Category != "" {
		expense.Category = input.Category
	}
	if input.Description != "" {
		expense.Description = input.Description
	}
	if input.Date != "" {
		if t, err := time.Parse("2006-01-02", input.Date); err == nil {
			expense.Date = t
		}
	}

	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, errors.New("failed to update expense")
	}
	return expense, nil
}

func (s *ExpenseService) DeleteExpense(userID, tripID, expenseID uuid.UUID) error {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return errors.New("trip not found")
	}
	if trip.UserID != userID {
		return errors.New("access denied")
	}

	expense, err := s.expenseRepo.FindByID(expenseID)
	if err != nil {
		return errors.New("expense not found")
	}
	if expense.TripID != tripID {
		return errors.New("expense does not belong to this trip")
	}

	return s.expenseRepo.Delete(expenseID)
}
