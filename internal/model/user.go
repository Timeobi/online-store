package model

import "time"

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User представляет пользователя системы.
// Обрати внимание: у структуры НЕТ прямого JSON-тега для PasswordHash с "-" —
// это специально, чтобы хэш пароля никогда случайно не утёк в API-ответах.
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
