package recipe

import "context"

type RecipeUC interface {
	CreateRecipe(ctx context.Context, recipe *RecipeModel) error
	GetRecipeByID(ctx context.Context, id string) (*RecipeModel, error)
	DeleteRecipe(ctx context.Context, id string) error
	UpdateRecipe(ctx context.Context, id string, dto UpdateRecipeDTO) error

	SearchRecipe(ctx context.Context, query string, tags []string, page, limit int64) ([]*RecipeModel, int, error)
}
