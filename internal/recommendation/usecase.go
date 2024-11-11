package recommendation

import "context"

type RecommendationUC interface {
	NewInteraction(ctx context.Context, userID, recipeID string, interaction int) error
	GetRecommendationCollaborative(ctx context.Context, userID string) ([]RecipeModel, error)
	GetRecommendationPreferences(ctx context.Context, userID string) ([]RecipeModel, error)
}
