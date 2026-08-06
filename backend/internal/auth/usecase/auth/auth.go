package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/AvitoTeam49/ScamTrainer/backend/internal/auth/entity"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type authRepository interface {
	CreateUser(ctx context.Context, email string, passwordHash string, role string) (user entity.User, err error)
	GetUserByEmail(ctx context.Context, email string) (user entity.User, err error)
}

type authService struct {
	authRepository authRepository
	logger         *zap.Logger
	jwtSecret      []byte
}

func NewAuthService(authRepository authRepository) *authService {
	return &authService{
		authRepository: authRepository,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (int64, string, error) {
	if err := validateRegisterInput(email, password); err != nil {
		return 0, "", err
	}

	_, err := s.authRepository.GetUserByEmail(ctx, email)

	if err == nil {
		return 0, "", entity.ErrUserAlreadyExists
	}

	if !errors.Is(err, entity.ErrUserNotFound) {
		return 0, "", fmt.Errorf("check user existence: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.authRepository.CreateUser(ctx, email, string(hashedPassword), "")
	if err != nil {
		return 0, "", fmt.Errorf("create user in repo: %w", err)
	}

	return user.ID, "user successfully registered", nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	if err := validateRegisterInput(email, password); err != nil {
		return "", "", entity.ErrInvalidCredentials
	}

	user, err := s.authRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return "", "", entity.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("find user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", "", entity.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("compare password hash: %w", err)
	}

	accessToken, err := s.generateToken(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(user.ID, user.Role, 7*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil

}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (bool, int64, string, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return false, 0, "", entity.ErrInvalidToken
	}

	return true, claims.UserID, claims.Role, nil
}

func validateRegisterInput(email, password string) error {
	if strings.TrimSpace(email) == "" {
		return entity.ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return entity.ErrInvalidEmail
	}

	if len([]rune(password)) < 8 {
		return entity.ErrInvalidPassword
	}

	return nil
}

func (s *authService) generateToken(userID int64, role string, ttl time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
