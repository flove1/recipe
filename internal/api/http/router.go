package http

import (
	"flove/job/config"
	"flove/job/internal/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	_ "flove/job/docs"
)

func newRouter(_ *config.Config, h Handlers) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/healthcheck", h.TokenHandler.RequireRole(user.RoleUser), healthcheck)

	r.POST("/users", h.UserHandler.CreateUser)
	r.GET("/users", h.TokenHandler.RequireAuthenticatedUser(), h.UserHandler.GetUserInfo)
	r.DELETE("/users", h.TokenHandler.RequireAuthenticatedUser(), h.UserHandler.DeleteUser)
	r.PATCH("/users", h.TokenHandler.RequireAuthenticatedUser(), h.UserHandler.UpdateUser)
	r.PATCH("/users/password", h.TokenHandler.RequireAuthenticatedUser(), h.UserHandler.ChangePassword)

	r.PATCH("/admin/users/role/:id", h.TokenHandler.RequireRole(user.RoleAdmin), h.UserHandler.ChangeUserRole)

	r.POST("/auth/sign-in", h.TokenHandler.SignIn)
	r.POST("/auth/sign-out", h.TokenHandler.SignOut)

	r.GET("/recommendations/collaborative", h.TokenHandler.RequireAuthenticatedUser(), h.RecommendationHandler.GetRecommendationCollaborative)
	r.GET("/recommendations/preferences", h.TokenHandler.RequireAuthenticatedUser(), h.RecommendationHandler.GetRecommendationByPreferences)
	r.POST("/recommendations/interaction", h.TokenHandler.RequireAuthenticatedUser(), h.RecommendationHandler.NewInteraction)

	r.POST("/recipes", h.TokenHandler.RequireRole(user.RoleAdmin), h.RecipeHandler.CreateRecipe)
	r.GET("/recipes", h.TokenHandler.RequireAuthenticatedUser(), h.RecipeHandler.SearchRecipe)
	r.GET("/recipes/:id", h.TokenHandler.RequireAuthenticatedUser(), h.RecipeHandler.GetRecipeByID)
	r.PATCH("/recipes/:id", h.TokenHandler.RequireRole(user.RoleAdmin), h.RecipeHandler.UpdateRecipe)
	r.DELETE("/recipes/:id", h.TokenHandler.RequireRole(user.RoleAdmin), h.RecipeHandler.DeleteRecipe)

	return r
}

func healthcheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "OK",
	})
}
