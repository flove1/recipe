package http

import (
	"flove/job/internal/auth"
	"flove/job/internal/recipe"
	"flove/job/internal/recommendation"
	"flove/job/internal/user"
)

type Handlers struct {
	UserHandler           *user.UserHandler
	TokenHandler          *auth.AuthHandler
	RecipeHandler         *recipe.RecipeHandler
	RecommendationHandler *recommendation.RecommendationHandler
}
