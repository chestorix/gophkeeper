package models

import "time"

type DataType string

const (
	LoginPassword DataType = "login_password"
	TextData      DataType = "text_data"
	BinaryData    DataType = "binary_data"
	CardData      DataType = "card_data"
)

type SecretItemData struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      DataType  `json:"type"`
	Name      string    `json:"name"`
	Metadata  string    `json:"metadata"`
	Data      []byte    `json:"data"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"create_at"`
	UpdatedAt    time.Time `json:"update_at"`
}

type LoginPasswordData struct {
	Login    string `json:"login"`
	Pass     string `json:"pass"`
	Resource string `json:"resource"`
}

type BankData struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Cardholder string `json:"cardholder"`
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type SyncRequest struct {
	LastSync time.Time        `json:"last_sync"`
	Data     []SecretItemData `json:"data"`
}
type SyncResponse struct {
	LastSync time.Time        `json:"last_sync"`
	Data     []SecretItemData `json:"data"`
}
