package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type MemberService struct {
	memberRepo *repository.MemberRepository
	tripRepo   *repository.TripRepository
	userRepo   *repository.UserRepository
}

func NewMemberService(memberRepo *repository.MemberRepository, tripRepo *repository.TripRepository, userRepo *repository.UserRepository) *MemberService {
	return &MemberService{
		memberRepo: memberRepo,
		tripRepo:   tripRepo,
		userRepo:   userRepo,
	}
}

type AddMemberInput struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=editor viewer"`
}

type UpdateMemberRoleInput struct {
	Role string `json:"role" binding:"required,oneof=editor viewer"`
}

func (s *MemberService) AddMember(ownerID, tripID uuid.UUID, input AddMemberInput) (*models.TripMember, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != ownerID {
		return nil, errors.New("only the trip owner can add members")
	}

	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.ID == ownerID {
		return nil, errors.New("cannot add yourself as a member")
	}

	if existing, _ := s.memberRepo.Find(tripID, user.ID); existing.UserID != uuid.Nil {
		return nil, errors.New("user is already a member of this trip")
	}

	member := &models.TripMember{
		TripID: tripID,
		UserID: user.ID,
		Role:   input.Role,
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, errors.New("failed to add member")
	}
	return member, nil
}

func (s *MemberService) ListMembers(userID, tripID uuid.UUID) ([]models.TripMember, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		if _, err := s.memberRepo.Find(tripID, userID); err != nil {
			return nil, errors.New("access denied")
		}
	}

	return s.memberRepo.ListByTripID(tripID)
}

func (s *MemberService) UpdateMemberRole(ownerID, tripID, memberUserID uuid.UUID, input UpdateMemberRoleInput) error {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return errors.New("trip not found")
	}
	if trip.UserID != ownerID {
		return errors.New("only the trip owner can update member roles")
	}

	if _, err := s.memberRepo.Find(tripID, memberUserID); err != nil {
		return errors.New("member not found")
	}

	return s.memberRepo.UpdateRole(tripID, memberUserID, input.Role)
}

func (s *MemberService) RemoveMember(ownerID, tripID, memberUserID uuid.UUID) error {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return errors.New("trip not found")
	}
	if trip.UserID != ownerID {
		return errors.New("only the trip owner can remove members")
	}

	if _, err := s.memberRepo.Find(tripID, memberUserID); err != nil {
		return errors.New("member not found")
	}

	return s.memberRepo.Delete(tripID, memberUserID)
}
