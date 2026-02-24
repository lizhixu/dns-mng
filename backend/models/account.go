package models

import "time"

type Account struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"`
	APIKey       string    `json:"api_key"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	Name         string `json:"name" binding:"required"`
	ProviderType string `json:"provider_type" binding:"required"`
	APIKey       string `json:"api_key" binding:"required"`
}

type UpdateAccountRequest struct {
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
}
