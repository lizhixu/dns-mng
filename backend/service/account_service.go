package service

import (
	"dns-mng/database"
	"dns-mng/models"
	"errors"
	"time"
)

type AccountService struct{}

func NewAccountService() *AccountService {
	return &AccountService{}
}

func (s *AccountService) List(userID int64) ([]models.Account, error) {
	rows, err := database.DB.Query(
		"SELECT id, user_id, name, provider_type, api_key, created_at, updated_at FROM accounts WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.ProviderType, &a.APIKey, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (s *AccountService) Get(userID, accountID int64) (*models.Account, error) {
	var a models.Account
	err := database.DB.QueryRow(
		"SELECT id, user_id, name, provider_type, api_key, created_at, updated_at FROM accounts WHERE id = ? AND user_id = ?",
		accountID, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.ProviderType, &a.APIKey, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AccountService) Create(userID int64, req *models.CreateAccountRequest) (*models.Account, error) {
	now := time.Now()
	result, err := database.DB.Exec(
		"INSERT INTO accounts (user_id, name, provider_type, api_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		userID, req.Name, req.ProviderType, req.APIKey, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &models.Account{
		ID:           id,
		UserID:       userID,
		Name:         req.Name,
		ProviderType: req.ProviderType,
		APIKey:       req.APIKey,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *AccountService) Update(userID, accountID int64, req *models.UpdateAccountRequest) (*models.Account, error) {
	// Check ownership
	account, err := s.Get(userID, accountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.APIKey != "" {
		account.APIKey = req.APIKey
	}
	account.UpdatedAt = time.Now()

	_, err = database.DB.Exec(
		"UPDATE accounts SET name = ?, api_key = ?, updated_at = ? WHERE id = ? AND user_id = ?",
		account.Name, account.APIKey, account.UpdatedAt, accountID, userID,
	)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AccountService) Delete(userID, accountID int64) error {
	result, err := database.DB.Exec("DELETE FROM accounts WHERE id = ? AND user_id = ?", accountID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("account not found")
	}
	return nil
}
