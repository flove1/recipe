package auth

import (
	"context"
	"flove/job/internal/user"
)

type RefreshTokenRepository interface {
	NewRefreshToken(ctx context.Context, token *RefreshTokenModel) error
	GetByToken(ctx context.Context, plaintext string) (*RefreshTokenModel, error)
	DeleteToken(ctx context.Context, token string) error
}

type AccessTokenRepository interface {
	NewAccessToken(ctx context.Context, token AccessTokenModel) error
	VerifyToken(ctx context.Context, token string) (string, user.Role, error)
}
