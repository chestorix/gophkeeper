// Package models содержит основные структуры данных для GophKeeper.
// Включает модели пользователей, секретных данных и запросов/ответов API.
package models

import "time"

// DataType представляет тип хранимых секретных данных.
type DataType string

const (
	// LoginPassword - тип для хранения пар логин/пароль.
	LoginPassword DataType = "login_password"

	// TextData - тип для хранения произвольных текстовых данных.
	TextData DataType = "text_data"

	// BinaryData - тип для хранения бинарных данных (файлов).
	BinaryData DataType = "binary_data"

	// CardData - тип для хранения данных банковских карт.
	CardData DataType = "card_data"
)

// SecretItemData представляет одну запись секретных данных пользователя.
// Содержит метаинформацию и сами данные в зашифрованном виде.
type SecretItemData struct {
	ID        string    `json:"id"`         // Уникальный идентификатор записи
	UserID    string    `json:"user_id"`    // ID владельца записи
	Type      DataType  `json:"type"`       // Тип хранимых данных
	Name      string    `json:"name"`       // Человеко-читаемое название
	Metadata  string    `json:"metadata"`   // Дополнительная метаинформация
	Data      []byte    `json:"data"`       // Данные в сериализованном виде
	Version   int       `json:"version"`    // Версия для разрешения конфликтов
	CreatedAt time.Time `json:"created_at"` // Время создания записи
	UpdatedAt time.Time `json:"updated_at"` // Время последнего обновления
}

// User представляет пользователя системы.
type User struct {
	ID           string    `json:"id"`            // Уникальный идентификатор пользователя
	Login        string    `json:"login"`         // Логин для входа в систему
	PasswordHash string    `json:"password_hash"` // Хэш пароля
	CreatedAt    time.Time `json:"create_at"`     // Время регистрации
	UpdatedAt    time.Time `json:"update_at"`     // Время последнего обновления
}

// LoginPasswordData содержит данные для типа LoginPassword.
type LoginPasswordData struct {
	Login    string `json:"login"`    // Логин
	Password string `json:"pass"`     // Пароль
	Resource string `json:"resource"` // Ресурс (сайт, приложение)
}

// BankData содержит данные банковской карты для типа CardData.
type BankData struct {
	CardNumber string `json:"card_number"` // Номер карты
	ExpiryDate string `json:"expiry_date"` // Срок действия
	CVV        string `json:"cvv"`         // CVV код
	Cardholder string `json:"cardholder"`  // Держатель карты
}

// AuthRequest представляет запрос на аутентификацию или регистрацию.
type AuthRequest struct {
	Login    string `json:"login"`    // Логин пользователя
	Password string `json:"password"` // Пароль пользователя
}

// AuthResponse представляет ответ с JWT токеном после успешной аутентификации.
type AuthResponse struct {
	Token string `json:"token"` // JWT токен для доступа к API
}

// SyncRequest представляет запрос на синхронизацию данных.
type SyncRequest struct {
	LastSync time.Time        `json:"last_sync"` // Время последней синхронизации
	Data     []SecretItemData `json:"data"`      // Данные для отправки на сервер
}

// SyncResponse представляет ответ на запрос синхронизации.
type SyncResponse struct {
	LastSync time.Time        `json:"last_sync"` // Текущее время сервера
	Data     []SecretItemData `json:"data"`      // Данные с сервера
}
