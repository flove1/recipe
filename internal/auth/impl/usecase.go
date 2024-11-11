package impl

import (
	"context"
	"flove/job/config"
	"flove/job/internal/auth"
	"flove/job/internal/base/database"
	"flove/job/internal/user"
	"time"
)

type useCase struct {
	cfg                    *config.Config
	accessTokenRepository  auth.AccessTokenRepository
	refreshTokenRepository auth.RefreshTokenRepository
	userRepository         user.UserRepository
}

func NewTokenUC(cfg *config.Config, accessTokenRepository auth.AccessTokenRepository, refreshTokenRepository auth.RefreshTokenRepository, userRepository user.UserRepository) auth.TokenUC {
	return &useCase{
		cfg:                    cfg,
		accessTokenRepository:  accessTokenRepository,
		refreshTokenRepository: refreshTokenRepository,
		userRepository:         userRepository,
	}
}

func (uc *useCase) NewRefreshToken(ctx context.Context, credentials string, password string) (*auth.RefreshTokenModel, error) {
	user, err := uc.userRepository.GetUserByCredentials(ctx, credentials)
	if err != nil {
		return nil, err
	}

	err = user.ComparePassword(password)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.NewRefreshToken(user.ID, time.Hour*24)
	if err != nil {
		return nil, err
	}

	err = uc.refreshTokenRepository.NewRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (uc *useCase) DeleteRefreshToken(ctx context.Context, token string) error {
	err := uc.refreshTokenRepository.DeleteToken(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) NewAccessToken(ctx context.Context, token string) (*auth.AccessTokenModel, error) {
	refreshToken, err := uc.refreshTokenRepository.GetByToken(ctx, token)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, auth.ErrInvalidToken
		}
		return nil, err
	}

	user, err := uc.userRepository.GetUserByID(ctx, refreshToken.UserUUID)
	if err != nil {
		return nil, err
	}

	accessToken, err := auth.NewAccessToken(refreshToken.UserUUID, user.Role)
	if err != nil {
		return nil, err
	}

	err = uc.accessTokenRepository.NewAccessToken(ctx, *accessToken)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func (uc *useCase) VerifyAccessToken(ctx context.Context, token string) (string, user.Role, error) {
	user, role, err := uc.accessTokenRepository.VerifyToken(ctx, token)
	if err != nil {
		if err == database.ErrNotFound {
			return "", 0, auth.ErrInvalidToken
		}
		return "", 0, err
	}

	return user, role, nil
}
