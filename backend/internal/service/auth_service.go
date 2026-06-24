package service

import (
	"context"
	"errors"
	"time"

	"morepark/internal/domain"
	"morepark/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// LoginResponse содержит токен и информацию о пользователе
type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// Login проверяет учётные данные и возвращает JWT
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// 1. Находим пользователя
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 2. Проверяем, активен ли пользователь
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// 3. Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 4. Генерируем JWT
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	// 5. Убираем хеш пароля из ответа
	user.PasswordHash = ""

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// generateJWT создаёт JWT токен
func (s *AuthService) generateJWT(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // токен живёт 24 часа
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken проверяет JWT и возвращает claims
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}
