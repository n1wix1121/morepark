package service

import (
	"context"
	"errors"

	"morepark/internal/domain"
	"morepark/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrUserEmailTaken     = errors.New("email уже занят")
	ErrInvalidRole        = errors.New("недопустимая роль")
	ErrCannotDeleteSelf   = errors.New("нельзя удалить свой аккаунт")
	ErrLastDirector       = errors.New("нельзя удалить последнего директора")
)

var validRoles = map[string]bool{
	"director":   true,
	"cashier":    true,
	"lifeguard":  true,
	"technician": true,
	"barman":     true,
}

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

func (s *UserService) GetAll(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*domain.User, error) {
	if !validRoles[req.Role] {
		return nil, ErrInvalidRole
	}
	if req.Password == "" {
		return nil, errors.New("пароль обязателен")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     req.IsActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, ErrUserEmailTaken
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (*domain.User, error) {
	if !validRoles[req.Role] {
		return nil, ErrInvalidRole
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user.Email = req.Email
	user.FullName = req.FullName
	user.Role = req.Role
	user.IsActive = req.IsActive

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hash)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id, currentUserID string) error {
	if id == currentUserID {
		return ErrCannotDeleteSelf
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Role == "director" {
		count, err := s.userRepo.CountByRole(ctx, "director")
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastDirector
		}
	}

	return s.userRepo.Delete(ctx, id)
}
