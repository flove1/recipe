package recipe

import (
	"context"
)

type RecipeRepository interface {
	CreateRecipe(ctx context.Context, recipe *RecipeModel) error
	GetRecipeByID(ctx context.Context, id string) (*RecipeModel, error)
	UpdateRecipe(ctx context.Context, id string, update UpdateRecipeDTO) error
	DeleteRecipe(ctx context.Context, id string) error

	SearchRecipe(ctx context.Context, query string, tags []string, page, limit int64) ([]*RecipeModel, int, error)
}
