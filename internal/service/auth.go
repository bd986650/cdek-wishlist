package service

import (
	"context"
	"errors"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/danil/cdek-wishlist/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Register(ctx context.Context, req model.RegisterRequest) (*model.TokenResponse, error) {
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, model.ErrAlreadyExists
	} else if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return nil, model.ErrAlreadyExists
		}
		return nil, err
	}

	tok, err := s.issueToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{Token: tok}, nil
}

func (s *authService) Login(ctx context.Context, req model.LoginRequest) (*model.TokenResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	tok, err := s.issueToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &model.TokenResponse{Token: tok}, nil
}

func (s *authService) issueToken(userID int64) (string, error) {
	now := time.Now().UTC()

	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     now.Add(s.jwtExpiry).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}
