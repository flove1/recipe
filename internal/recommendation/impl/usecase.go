package impl

import (
	"context"
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/recommendation"
)

type usecase struct {
	config             *config.Config
	eventBus           *database.EventBus
	recommendationRepo recommendation.RecommendationRepository
}

func NewRecommendationUC(config *config.Config, eventBus *database.EventBus, repo recommendation.RecommendationRepository) recommendation.RecommendationUC {
	return &usecase{
		config:             config,
		eventBus:           eventBus,
		recommendationRepo: repo,
	}
}

func (u *usecase) GetRecommendationCollaborative(ctx context.Context, userID string) ([]recommendation.RecipeModel, error) {
	recipes, err := u.recommendationRepo.GetRecommendationCollaborative(ctx, userID)
	if err != nil {
		return nil, err
	}

	return recipes, nil
}

func (u *usecase) GetRecommendationPreferences(ctx context.Context, userID string) ([]recommendation.RecipeModel, error) {
	recipes, err := u.recommendationRepo.GetRecommendationPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	return recipes, nil
}

func (u *usecase) NewInteraction(ctx context.Context, userID string, recipeID string, interaction int) error {
	err := u.recommendationRepo.NewInteraction(ctx, userID, recipeID, interaction)
	if err != nil {
		return err
	}

	err = u.recommendationRepo.RecalculatePreferences(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
