package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"dns-mng/database"
	"dns-mng/models"
)

type DDNSTokenService struct{}

func NewDDNSTokenService() *DDNSTokenService {
	return &DDNSTokenService{}
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GetTokenByValue retrieves a token by its value
func (s *DDNSTokenService) GetTokenByValue(tokenValue string) (*models.DDNSToken, error) {
	var token models.DDNSToken
	var enabled int
	err := database.DB.QueryRow(`
		SELECT id, user_id, token, enabled, 
			COALESCE(last_used_at, '') as last_used_at, 
			COALESCE(last_ip, '') as last_ip, 
			created_at, updated_at
		FROM ddns_tokens
		WHERE token = ?
	`, tokenValue).Scan(
		&token.ID, &token.UserID, &token.Token, &enabled,
		&token.LastUsedAt, &token.LastIP, &token.CreatedAt, &token.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	token.Enabled = enabled == 1
	return &token, nil
}

// GetToken retrieves a token for a user
func (s *DDNSTokenService) GetToken(userID int64) (*models.DDNSToken, error) {
	var token models.DDNSToken
	var enabled int
	err := database.DB.QueryRow(`
		SELECT id, user_id, token, enabled, 
			COALESCE(last_used_at, '') as last_used_at, 
			COALESCE(last_ip, '') as last_ip, 
			created_at, updated_at
		FROM ddns_tokens
		WHERE user_id = ?
	`, userID).Scan(
		&token.ID, &token.UserID, &token.Token, &enabled,
		&token.LastUsedAt, &token.LastIP, &token.CreatedAt, &token.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	token.Enabled = enabled == 1
	return &token, nil
}

// CreateToken creates a new token for a user
func (s *DDNSTokenService) CreateToken(userID int64, customToken string) (*models.DDNSToken, error) {
	token := customToken
	if token == "" {
		token = generateRandomToken()
	}

	result, err := database.DB.Exec(`
		INSERT INTO ddns_tokens (user_id, token, enabled, created_at, updated_at)
		VALUES (?, ?, 1, datetime('now'), datetime('now'))
	`, userID, token)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.DDNSToken{
		ID:        id,
		UserID:    userID,
		Token:     token,
		Enabled:   true,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateToken updates an existing token
func (s *DDNSTokenService) UpdateToken(userID int64, enabled *bool, customToken string) (*models.DDNSToken, error) {
	var token string
	var id int64
	var currentEnabled int

	err := database.DB.QueryRow(`
		SELECT id, token, enabled FROM ddns_tokens WHERE user_id = ?
	`, userID).Scan(&id, &token, &currentEnabled)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if customToken != "" {
		token = customToken
	}

	enabledVal := currentEnabled
	if enabled != nil {
		enabledVal = boolToInt(*enabled)
	}

	_, err = database.DB.Exec(`
		UPDATE ddns_tokens
		SET token = ?, enabled = ?, updated_at = datetime('now')
		WHERE user_id = ?
	`, token, enabledVal, userID)
	if err != nil {
		return nil, err
	}

	return &models.DDNSToken{
		ID:        id,
		UserID:    userID,
		Token:     token,
		Enabled:   enabledVal == 1,
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteToken deletes a token for a user
func (s *DDNSTokenService) DeleteToken(userID int64) error {
	_, err := database.DB.Exec(`DELETE FROM ddns_tokens WHERE user_id = ?`, userID)
	return err
}

// UpdateLastUsed updates the last used timestamp and IP
func (s *DDNSTokenService) UpdateLastUsed(tokenValue string, ip string) error {
	_, err := database.DB.Exec(`
		UPDATE ddns_tokens
		SET last_used_at = datetime('now'), last_ip = ?, updated_at = datetime('now')
		WHERE token = ?
	`, ip, tokenValue)
	return err
}
