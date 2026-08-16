package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserType string

const (
	UserTypePassenger UserType = "passenger"
	UserTypeDriver    UserType = "driver"
)

var (
	ErrUserNotFound      = errors.New("usuário não encontrado")
	ErrUserAlreadyExists = errors.New("email de usuário já cadastrado")
	ErrInvalidUserType   = errors.New("tipo de usuário inválido (deve ser 'passenger' ou 'driver')")
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UserType  UserType  `json:"user_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
