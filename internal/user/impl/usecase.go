package impl

import (
	"context"
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/user"
)

type useCase struct {
	cfg      *config.Config
	eventBus *database.EventBus
	userRepo user.UserRepository
}

func NewUserUC(config *config.Config, eventBus *database.EventBus, userRepository user.UserRepository) user.UserUC {
	return &useCase{
		cfg:      config,
		eventBus: eventBus,
		userRepo: userRepository,
	}
}

func (uc *useCase) CreateUser(ctx context.Context, user *user.UserModel) error {
	if err := uc.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	if err := uc.eventBus.Publish("user:created", user.ID); err != nil {
		return err
	}

	return nil
}

func (uc *useCase) UpdateUser(ctx context.Context, userID string, updates any) error {
	return uc.userRepo.UpdateUser(ctx, userID, updates)
}

func (uc *useCase) ChangePassword(ctx context.Context, id string, password string) error {
	model, err := uc.userRepo.GetUserByID(ctx, id)
	model.SetPassword(password)

	if err != nil {
		return err
	}

	return nil
}

func (uc *useCase) GetUserByID(ctx context.Context, id string) (*user.UserModel, error) {
	return uc.userRepo.GetUserByID(ctx, id)
}

func (uc *useCase) DeleteUser(ctx context.Context, userID string) error {
	if err := uc.userRepo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	if err := uc.eventBus.Publish("user:deleted", userID); err != nil {
		return err
	}

	return nil
}

func (uc *useCase) ChangeUserRole(ctx context.Context, userID string, role user.Role) error {
	return uc.userRepo.ChangeUserRole(ctx, userID, role)
}
