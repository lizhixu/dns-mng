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

func (s *UserService) GetUser(userID int64) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) UpdatePassword(userID int64, req *models.UpdatePasswordRequest) error {
	// Get current password hash
	var currentHash string
	err := database.DB.QueryRow(
		"SELECT password_hash FROM users WHERE id = ?",
		userID,
	).Scan(&currentHash)

	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	_, err = database.DB.Exec(
		"UPDATE users SET password_hash = ? WHERE id = ?",
		string(newHash), userID,
	)

	return err
}
