package recipe

import (
	"flove/job/config"
	"flove/job/internal/base/database"
	"flove/job/internal/base/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RecipeHandler struct {
	config   *config.Config
	recipeUC RecipeUC
}

func NewRecipeHandler(config *config.Config, uc RecipeUC) *RecipeHandler {
	return &RecipeHandler{
		config:   config,
		recipeUC: uc,
	}
}

type nutrition struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Fat           float64 `json:"fat"`
	Carbohydrates float64 `json:"carbohydrates"`
	Fiber         float64 `json:"fiber"`
	Sugar         float64 `json:"sugar"`
	Sodium        float64 `json:"sodium"`
}

type createRecipeRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Category    string    `json:"category" binding:"required"`
	Tags        []string  `json:"tags"`
	Nutrition   nutrition `json:"nutrition"`
	Servings    int       `json:"servings" binding:"required"`
}

// @Summary Create a new recipe
// @Description Create a new recipe with the given details
// @Security BasicAuth
// @Tags Recipe
// @Accept json
// @Produce json
// @Param recipe body createRecipeRequest true "Recipe details"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recipes [post]
func (h *RecipeHandler) CreateRecipe(ctx *gin.Context) {
	var req createRecipeRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	model := &RecipeModel{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Nutrition:   NutritionInfo(req.Nutrition),
		Servings:    req.Servings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.recipeUC.CreateRecipe(ctx, model); err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteResponse(ctx, http.StatusCreated, "recipe succesfully created")
}

// @Summary Delete a recipe
// @Description Delete a recipe with the given ID
// @Security BasicAuth
// @Tags Recipe
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recipes/{id} [delete]
func (h *RecipeHandler) DeleteRecipe(ctx *gin.Context) {
	var req struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.recipeUC.DeleteRecipe(ctx, req.ID); err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "recipe succesfully deleted")
}

// @Summary Get a recipe by ID
// @Description Get a recipe by the given ID
// @Security BasicAuth
// @Tags Recipe
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recipes/{id} [get]
func (h *RecipeHandler) GetRecipeByID(ctx *gin.Context) {
	var req struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	recipe, err := h.recipeUC.GetRecipeByID(ctx, req.ID)
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Code: http.StatusOK,
		Body: recipe,
	})
}

type updateRecipeRequest struct {
	ID          string   `uri:"id" binding:"required"`
	Name        string   `json:"name" binding:"omitempty"`
	Description string   `json:"description" binding:"omitempty"`
	Category    string   `json:"category" binding:"omitempty"`
	Tags        []string `json:"tags" binding:"omitempty"`
	CookTime    int      `json:"cook_time" binding:"omitempty"`
	Servings    int      `json:"servings" binding:"omitempty"`
}

// @Summary Update a recipe
// @Description Update a recipe with the given ID
// @Security BasicAuth
// @Tags Recipe
// @Accept json
// @Produce json
// @Param id path string true "Recipe ID"
// @Param recipe body updateRecipeRequest true "Recipe details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recipes/{id} [patch]
func (h *RecipeHandler) UpdateRecipe(ctx *gin.Context) {
	var req updateRecipeRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID := ctx.MustGet("userID").(string)
	err := h.recipeUC.UpdateRecipe(ctx, userID, UpdateRecipeDTO{})
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, &response.Response{
		Code:    http.StatusOK,
		Message: "user succesfully updated",
	})
}

type searchParametersRequest struct {
	Query string   `json:"query" binding:"omitempty"`
	Tags  []string `json:"tags" binding:"omitempty"`
	Page  int64    `json:"page" binding:"omitempty"`
	Limit int64    `json:"limit" binding:"omitempty"`
}

// @Summary Search recipes
// @Description Search recipes based on query, tags, page, and limit
// @Security BasicAuth
// @Tags Recipe
// @Accept json
// @Produce json
//
//	@Param searchRequest body searchParametersRequest true "Search parameters"
//
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recipes/search [get]
func (h *RecipeHandler) SearchRecipe(ctx *gin.Context) {
	var req searchParametersRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	recipes, totalDocuments, err := h.recipeUC.SearchRecipe(ctx, req.Query, req.Tags, req.Page, req.Limit)
	if err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, response.Response{
		Code: http.StatusOK,
		Body: struct {
			Recipes []*RecipeModel `json:"recipes"`
			Total   int            `json:"total"`
			Limit   int64          `json:"limit"`
			Page    int64          `json:"page"`
		}{
			Recipes: recipes,
			Total:   totalDocuments,
			Limit:   req.Limit,
			Page:    max(1, req.Page),
		},
	})
}
