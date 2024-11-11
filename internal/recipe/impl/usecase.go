package impl

import (
	"context"
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/recipe"
	"fmt"
	"strings"
)

type usecase struct {
	config     *config.Config
	eventBus   *database.EventBus
	recipeRepo recipe.RecipeRepository
}

func NewRecipeUC(config *config.Config, eventBus *database.EventBus, repo recipe.RecipeRepository) recipe.RecipeUC {
	return &usecase{
		config:     config,
		eventBus:   eventBus,
		recipeRepo: repo,
	}
}

func (uc *usecase) CreateRecipe(ctx context.Context, recipe *recipe.RecipeModel) error {
	if err := uc.recipeRepo.CreateRecipe(ctx, recipe); err != nil {
		return err
	}

	if err := uc.eventBus.Publish("recipe:created",
		fmt.Sprintf("%s:%s:%s:%s",
			recipe.ID,
			recipe.Name,
			recipe.Category,
			strings.Join(recipe.Tags, ","),
		)); err != nil {
		return err
	}

	return nil
}

func (uc *usecase) DeleteRecipe(ctx context.Context, id string) error {
	if err := uc.recipeRepo.DeleteRecipe(ctx, id); err != nil {
		return err
	}

	if err := uc.eventBus.Publish("recipe:deleted", id); err != nil {
		return err
	}

	return nil
}

func (uc *usecase) GetRecipeByID(ctx context.Context, id string) (*recipe.RecipeModel, error) {
	recipe, err := uc.recipeRepo.GetRecipeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return recipe, nil
}

func (uc *usecase) UpdateRecipe(ctx context.Context, id string, dto recipe.UpdateRecipeDTO) error {
	if err := uc.recipeRepo.UpdateRecipe(ctx, id, dto); err != nil {
		return err
	}

	return nil
}

func (uc *usecase) SearchRecipe(ctx context.Context, query string, tags []string, page, limit int64) ([]*recipe.RecipeModel, int, error) {
	recipes, totakDocuments, err := uc.recipeRepo.SearchRecipe(ctx, query, tags, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return recipes, totakDocuments, nil
}
