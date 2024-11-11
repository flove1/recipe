package user

import (
	"context"
)

type UserUC interface {
	CreateUser(ctx context.Context, user *UserModel) error
	UpdateUser(ctx context.Context, userID string, updates any) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, id string) (*UserModel, error)

	ChangeUserRole(ctx context.Context, userID string, role Role) error
	ChangePassword(ctx context.Context, uuid string, password string) error
}
