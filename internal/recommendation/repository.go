package recommendation

import "context"

type RecommendationRepository interface {
	NewInteraction(ctx context.Context, userID, recipeID string, interaction int) error
	RecalculatePreferences(ctx context.Context, userID string) error
	GetRecommendationCollaborative(ctx context.Context, userID string) ([]RecipeModel, error)
	GetRecommendationPreferences(ctx context.Context, userID string) ([]RecipeModel, error)
}
