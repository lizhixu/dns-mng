package service

import (
	"database/sql"
	"errors"
	"time"

	"dns-mng/config"
	"dns-mng/database"
	"dns-mng/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	cfg *config.Config
}

func NewUserService(cfg *config.Config) *UserService {
	return &UserService{cfg: cfg}
}

func (s *UserService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if username exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert user
	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		req.Username, string(hash),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	user := models.User{
		ID:        id,
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: token, User: user}, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	var user models.User
	var passwordHash string

	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &passwordHash, &user.CreatedAt)

	// 如果用户不存在，自动注册
	if err == sql.ErrNoRows {
		return s.Register(&models.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
		})
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: token, User: user}, nil
}

func (s *UserService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
