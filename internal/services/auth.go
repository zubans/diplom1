package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/dto"
	"gophermart/internal/storage/repos"
	"time"
)

type AuthService struct {
	userRepo  repos.UserRepository
	jwtSecret []byte
}

func NewAuthService(
	userRepo repos.UserRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req dto.CredentialsRequest,
) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return 0, err
	}

	return s.userRepo.CreateUser(ctx, req.Login, string(hashedPassword))
}

func (s *AuthService) Login(
	ctx context.Context,
	req dto.CredentialsRequest,
) (string, error) {
	storedHash, err := s.userRepo.GetPasswordHash(ctx, req.Login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(storedHash),
		[]byte(req.Password),
	); err != nil {
		return "", errors.New("invalid credentials")
	}

	user, err := s.userRepo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		return "", err
	}

	return s.GenerateToken(user.ID)
}

func (s *AuthService) GenerateToken(userID int) (string, error) {
	if userID == 0 {
		return "", errors.New("userID cannot be empty")
	}

	if len(s.jwtSecret) == 0 {
		return "", errors.New("JWT secret not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
