package services

import (
	"context"
	"errors"
	"fmt"
	"fullstack-simple-app/internal/errcode"
	"fullstack-simple-app/internal/models"
	"fullstack-simple-app/pkg/app_errors"
	"fullstack-simple-app/pkg/tokens/authentication"
	"fullstack-simple-app/pkg/tokens/verification"
	"fullstack-simple-app/pkg/validator"
	"log"
	"time"
)

type UserService struct {
	userRepository UserRepo
	userAdapter    EmailSender
	asyncRunner    AsyncRunner
	redisClient    RedisClient
	tokenMaker     TokenMaker
}

type AsyncRunner interface {
	RunAsync(fn func())
}

type UserRepo interface {
	CreateUser(user *models.User) error
	ActivateUser(email string) (models.User, error)
	GetUserIDByEmail(email string) (int64, error)
	GetUserByEmail(email string) (models.User, error)
}

type EmailSender interface {
	SendMail(recipient string, templateFile string, data interface{}) error
}

type RedisClient interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type TokenMaker interface {
	CreateToken(username string, duration time.Duration) (string, error)
	VerifyToken(token string) (*authentication.Payload, error)
}

func NewUserService(userRepo UserRepo, EmailSender EmailSender, async AsyncRunner, redis RedisClient, maker TokenMaker) *UserService {
	return &UserService{
		userRepository: userRepo,
		userAdapter:    EmailSender,
		asyncRunner:    async,
		redisClient:    redis,
		tokenMaker:     maker,
	}
}

func (s *UserService) RegisterUser(user *models.User, password string) error {
	const op = "RegisterUser"

	err := user.Password.Set(password)
	if err != nil {
		return fmt.Errorf("%s: user.Password.Set: %w", op, err)
	}

	v := validator.New()

	if models.ValidateUser(v, user); !v.Valid() {
		return app_errors.NewAppError(errcode.ErrInvalidRequest, fmt.Errorf("%v", v.Errors))
	}

	err = s.userRepository.CreateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			return app_errors.NewAppError(errcode.ErrEmailAlreadyExists, err)
		default:
			return app_errors.NewAppError(errcode.ErrInternal, err)
		}
	}

	otp, err := verification.GenerateOTP()
	if err != nil {
		return app_errors.NewAppError(errcode.ErrAccountCreated, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = s.redisClient.Set(ctx, "activation:"+user.Email, otp, 15*time.Minute)
	if err != nil {
		return app_errors.NewAppError(errcode.ErrAccountCreated, err)
	}

	s.asyncRunner.RunAsync(func() {
		data := map[string]interface{}{
			"activationToken": otp,
			"userID":          user.UserID,
		}
		err := s.userAdapter.SendMail(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			log.Printf("Failed to send confirmation email: %v\n", err)
		}
	})

	return nil
}

func (s *UserService) VerifyUser(email string, otp string) (models.User, string, error) {
	v := validator.New()

	if models.ValidateEmail(v, email); !v.Valid() {
		return models.User{}, "", app_errors.NewAppError(errcode.ErrInvalidRequest, fmt.Errorf("%v", v.Errors))
	}

	val, err := s.redisClient.Get(context.Background(), "activation:"+email)
	if err != nil {
		return models.User{}, "", app_errors.NewAppError(errcode.ErrOTPNotFound, err)
	}

	if val != otp {
		return models.User{}, "", app_errors.NewAppError(errcode.ErrOTPInvalid, errors.New("Invalid otp provided"))
	}

	s.asyncRunner.RunAsync(func() {
		err := s.redisClient.Del(context.Background(), "activation:"+email)
		if err != nil {
			log.Printf("Failed to send confirmation email: %v\n", err)
		}
	})

	user, err := s.userRepository.ActivateUser(email)
	if err != nil {
		return models.User{}, "", app_errors.NewAppError(errcode.ErrInternal, err)
	}

	token, err := s.tokenMaker.CreateToken(user.Email, 24*time.Hour)
	if err != nil {
		return models.User{}, "", app_errors.NewAppError(errcode.ErrLoginRedirect, err)
	}

	return user, token, nil
}

func (s *UserService) ResendCode(email string) error {
	v := validator.New()

	if models.ValidateEmail(v, email); !v.Valid() {
		return app_errors.NewAppError(errcode.ErrInvalidRequest, fmt.Errorf("%v", v.Errors))
	}

	otp, err := verification.GenerateOTP()
	if err != nil {
		return app_errors.NewAppError(errcode.ErrAccountCreated, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = s.redisClient.Set(ctx, "activation:"+email, otp, 15*time.Minute)
	if err != nil {
		return app_errors.NewAppError(errcode.ErrAccountCreated, err)
	}

	userID, err := s.userRepository.GetUserIDByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return app_errors.NewAppError(errcode.ErrNotFound, err)
		}
		return app_errors.NewAppError(errcode.ErrInternal, err)
	}

	s.asyncRunner.RunAsync(func() {
		data := map[string]interface{}{
			"activationToken": otp,
			"userID":          userID,
		}
		err := s.userAdapter.SendMail(email, "user_welcome.tmpl", data)
		if err != nil {
			log.Printf("Failed to send confirmation email: %v\n", err)
		}
	})

	return nil
}

func (s *UserService) UserSignIn(email string, password string) (string, error) {
	v := validator.New()

	models.ValidateEmail(v, email)
	models.ValidatePasswordPlaintext(v, password)

	if !v.Valid() {
		return "", app_errors.NewAppError(errcode.ErrInvalidRequest, fmt.Errorf("%v", v.Errors))
	}

	user, err := s.userRepository.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return "", app_errors.NewAppError(errcode.ErrNotFound, err)
		}
		return "", app_errors.NewAppError(errcode.ErrInternal, err)
	}

	match, err := user.Password.Matches(password)
	if err != nil {
		return "", app_errors.NewAppError(errcode.ErrInternal, err)
	}

	if !match {
		return "", app_errors.NewAppError(errcode.ErrInvalidPassword, errors.New("invalid password"))
	}

	return s.tokenMaker.CreateToken(user.Email, 24*time.Hour)
}

func (s *UserService) GetUser(email string) (models.User, error) {
	v := validator.New()

	models.ValidateEmail(v, email)

	if !v.Valid() {
		return models.User{}, app_errors.NewAppError(errcode.ErrInvalidRequest, fmt.Errorf("%v", v.Errors))
	}

	user, err := s.userRepository.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, app_errors.NewAppError(errcode.ErrNotFound, err)
		}
		return models.User{}, app_errors.NewAppError(errcode.ErrInternal, err)
	}

	return user, nil
}
