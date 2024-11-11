package auth

import (
	"context"
	"flove/job/internal/user"
)

type TokenUC interface {
	NewRefreshToken(ctx context.Context, credentials string, password string) (*RefreshTokenModel, error)
	DeleteRefreshToken(ctx context.Context, token string) error

	NewAccessToken(ctx context.Context, refreshToken string) (*AccessTokenModel, error)
	VerifyAccessToken(ctx context.Context, token string) (string, user.Role, error)
}
