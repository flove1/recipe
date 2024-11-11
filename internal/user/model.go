package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	ID           string
	Username     string
	Email        string
	Phone        string
	PasswordHash []byte
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role int

const (
	RoleUser Role = iota
	RoleAdmin
)

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (user *UserModel) ComparePassword(password string) error {
	err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return ErrMismatchedPassword
		default:
			return err
		}
	}

	return nil
}

func (user *UserModel) SetPassword(newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash

	return nil
}
