package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(expense *models.TripExpense) error {
	return r.db.Create(expense).Error
}

func (r *ExpenseRepository) FindByID(id uuid.UUID) (*models.TripExpense, error) {
	var expense models.TripExpense
	err := r.db.Where("id = ?", id).First(&expense).Error
	return &expense, err
}

func (r *ExpenseRepository) ListByTripID(tripID uuid.UUID, page, perPage int) ([]models.TripExpense, int64, error) {
	var expenses []models.TripExpense
	var total int64

	query := r.db.Model(&models.TripExpense{}).Where("trip_id = ?", tripID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order("date DESC").Find(&expenses).Error
	return expenses, total, err
}

func (r *ExpenseRepository) Update(expense *models.TripExpense) error {
	return r.db.Save(expense).Error
}

func (r *ExpenseRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.TripExpense{}).Error
}
