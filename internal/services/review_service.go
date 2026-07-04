package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	placeRepo  *repository.PlaceRepository
	awarder    PointsAwarder
}

func NewReviewService(reviewRepo *repository.ReviewRepository, placeRepo *repository.PlaceRepository) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		placeRepo:  placeRepo,
	}
}

// SetPointsAwarder wires the gamification awarder (optional).
func (s *ReviewService) SetPointsAwarder(a PointsAwarder) {
	s.awarder = a
}

type CreateReviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

type UpdateReviewInput struct {
	Rating  int    `json:"rating" binding:"min=0,max=5"`
	Comment string `json:"comment"`
}

func (s *ReviewService) CreateReview(userID uuid.UUID, placeID uint, input CreateReviewInput) (*models.PlaceReview, error) {
	if _, err := s.placeRepo.FindByID(placeID); err != nil {
		return nil, errors.New("place not found")
	}

	if existing, _ := s.reviewRepo.FindByUserAndPlace(userID, placeID); existing.ID != uuid.Nil {
		return nil, errors.New("you have already reviewed this place")
	}

	review := &models.PlaceReview{
		ID:      uuid.New(),
		PlaceID: placeID,
		UserID:  userID,
		Rating:  input.Rating,
		Comment: input.Comment,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, errors.New("failed to create review")
	}

	if s.awarder != nil {
		_ = s.awarder.AwardPoints(userID, PointsPerReview)
	}

	return review, nil
}

func (s *ReviewService) ListReviews(placeID uint, page, perPage int) ([]models.PlaceReview, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return s.reviewRepo.ListByPlaceID(placeID, page, perPage)
}

func (s *ReviewService) UpdateReview(reviewID uuid.UUID, userID uuid.UUID, input UpdateReviewInput) (*models.PlaceReview, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, errors.New("review not found")
	}
	if review.UserID != userID {
		return nil, errors.New("access denied")
	}

	if input.Rating > 0 {
		review.Rating = input.Rating
	}
	if input.Comment != "" {
		review.Comment = input.Comment
	}

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, errors.New("failed to update review")
	}
	return review, nil
}

func (s *ReviewService) DeleteReview(reviewID uuid.UUID, userID uuid.UUID) error {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return errors.New("review not found")
	}
	if review.UserID != userID {
		return errors.New("access denied")
	}
	return s.reviewRepo.Delete(reviewID)
}
