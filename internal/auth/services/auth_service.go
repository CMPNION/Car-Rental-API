package services

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	db        *gorm.DB
	validate  *validator.Validate
	jwtSecret []byte
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	Token string       `json:"token"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		validate:  validator.New(),
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(req RegisterRequest) (RegisterResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return RegisterResponse{}, err
	}

	var existing models.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return RegisterResponse{}, ErrEmailTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return RegisterResponse{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return RegisterResponse{}, err
	}

	user := models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.UserRoleClient,
		Balance:      0,
		Rating:       0,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return RegisterResponse{}, err
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{Token: token}, nil
}

func (s *AuthService) Login(req LoginRequest) (LoginResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return LoginResponse{}, err
	}

	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{Token: token}, nil
}

func (s *AuthService) generateToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
