package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"recon/internal/utils/helper"

	"golang.org/x/crypto/bcrypt"
	"recon/internal/config"
	"recon/internal/models"
	"recon/internal/repositories/user"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUsernameExists  = errors.New("username already exists")
	ErrInvalidUserData = errors.New("invalid user data")
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, user *models.User) error
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	ChangePassword(ctx context.Context, id string, newPassword string) error
}

type service struct {
	userRepo user.Repository
}

func NewService() Service {
	return &service{
		userRepo: user.NewRepository(),
	}
}

func (s *service) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	hashPwd := helper.Sha256Hex(password)
	if hashPwd != user.Password {
		return "", ErrInvalidPassword
	}

	now := time.Now()
	expire := now.Add(config.GlobalConfig.JWT.Expire)

	claims := jwt.MapClaims{
		"userID":   user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      expire.Unix(),
		"iat":      now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}

func (s *service) Register(ctx context.Context, user *models.User) error {

	if user.Username == "" || user.Password == "" || user.Email == "" {
		return ErrInvalidUserData
	}

	existingUser, err := s.userRepo.FindByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUsernameExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.userRepo.Create(ctx, user)
}

func (s *service) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *service) UpdateUser(ctx context.Context, id string, user *models.User) error {

	existingUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	return s.userRepo.Update(ctx, id, user)
}

func (s *service) DeleteUser(ctx context.Context, id string) error {

	existingUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *service) ChangePassword(ctx context.Context, id string, newPassword string) error {
	if newPassword == "" {
		return ErrInvalidUserData
	}
	hashed := helper.Sha256Hex(newPassword)
	fields := bson.M{
		"password":   hashed,
		"updated_at": time.Now(),
	}
	return s.userRepo.UpdateFields(ctx, id, fields)
}
