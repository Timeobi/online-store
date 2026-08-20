package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/repository"
)

var (
	ErrEmailAlreadyExists = errors.New("пользователь с таким email уже существует")
	ErrInvalidCredentials = errors.New("неверный email или пароль")
)

const tokenTTL = 24 * time.Hour

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

type RegisterInput struct {
	Email    string
	Password string
}

type AuthOutput struct {
	User  *model.User
	Token string
}

// Register создаёт нового пользователя с захэшированным паролем.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("некорректный email")
	}
	if len(input.Password) < 8 {
		return nil, errors.New("пароль должен быть не короче 8 символов")
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(ctx, email, hash)
	if err != nil {
		return nil, err
	}

	token, err := generateToken(user, s.jwtSecret, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{User: user, Token: token}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

// Login проверяет учётные данные и выдаёт новый токен.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Специально не различаем "пользователь не найден" и "неверный пароль" в ответе —
	// иначе злоумышленник может через простой перебор понять, какие email вообще
	// зарегистрированы в системе (так называемая "user enumeration" уязвимость).
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !checkPassword(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := generateToken(user, s.jwtSecret, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{User: user, Token: token}, nil
}
