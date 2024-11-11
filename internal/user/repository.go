package user

import "context"

type UserRepository interface {
	CreateUser(ctx context.Context, user *UserModel) error
	UpdateUser(ctx context.Context, userID string, updates any) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, id string) (*UserModel, error)
	GetUserByCredentials(ctx context.Context, email string) (*UserModel, error)

	ChangeUserRole(ctx context.Context, userID string, role Role) error
	ChangeUserPassword(ctx context.Context, userID string, password string) error
}
