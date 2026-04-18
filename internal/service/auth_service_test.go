package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danil/cdek-wishlist/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*model.User, error)
	createFn     func(ctx context.Context, user *model.User) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createFn == nil {
		return errors.New("createFn not set")
	}
	return m.createFn(ctx, user)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFn == nil {
		return nil, errors.New("getByEmailFn not set")
	}
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return nil, errors.New("not used in tests")
}

func TestAuthService_Register_AlreadyExistsPreCheck(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: 1, Email: email}, nil
		},
	}

	svc := NewAuthService(repo, "secret", 3600*time.Second)
	_, err := svc.Register(context.Background(), model.RegisterRequest{Email: "a@b.c", Password: "password12"})
	if !errors.Is(err, model.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAuthService_Register_GetByEmailUnexpectedError(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, errors.New("db down")
		},
	}

	svc := NewAuthService(repo, "secret", 3600*time.Second)
	_, err := svc.Register(context.Background(), model.RegisterRequest{Email: "a@b.c", Password: "password12"})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestAuthService_Register_CreateConflict(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		createFn: func(ctx context.Context, user *model.User) error {
			return model.ErrAlreadyExists
		},
	}

	svc := NewAuthService(repo, "secret", 3600*time.Second)
	_, err := svc.Register(context.Background(), model.RegisterRequest{Email: "a@b.c", Password: "password12"})
	if !errors.Is(err, model.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAuthService_Register_Success_IssuesJWT(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
		createFn: func(ctx context.Context, user *model.User) error {
			user.ID = 42
			user.CreatedAt = time.Now().UTC()
			return nil
		},
	}

	svc := NewAuthService(repo, "super-secret", 3600*time.Second)
	resp, err := svc.Register(context.Background(), model.RegisterRequest{Email: "a@b.c", Password: "password12"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil || resp.Token == "" {
		t.Fatalf("expected token")
	}

	token, err := jwt.Parse(resp.Token, func(t *jwt.Token) (interface{}, error) {
		return []byte("super-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("invalid jwt: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected map claims")
	}
	if int64(claims["user_id"].(float64)) != 42 {
		t.Fatalf("unexpected user_id claim: %#v", claims["user_id"])
	}
}

func TestAuthService_Login_InvalidCredentials_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, model.ErrNotFound
		},
	}
	svc := NewAuthService(repo, "secret", 3600*time.Second)
	_, err := svc.Login(context.Background(), model.LoginRequest{Email: "a@b.c", Password: "x"})
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_InvalidCredentials_BadPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: 7, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewAuthService(repo, "secret", 3600*time.Second)
	_, err = svc.Login(context.Background(), model.LoginRequest{Email: "a@b.c", Password: "wrong"})
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: 9, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewAuthService(repo, "secret", 3600*time.Second)
	resp, err := svc.Login(context.Background(), model.LoginRequest{Email: "a@b.c", Password: "right"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil || resp.Token == "" {
		t.Fatalf("expected token")
	}
}
