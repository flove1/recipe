package recommendation

import (
	"flove/job/config"
	"flove/job/internal/base/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	cfg              *config.Config
	recommendationUC RecommendationUC
}

func NewRecommendationHandler(cfg *config.Config, uc RecommendationUC) *RecommendationHandler {
	return &RecommendationHandler{
		cfg:              cfg,
		recommendationUC: uc,
	}
}

// @Summary Get recommendation by similar users
// @Description Get recommendation based on similar users
// @Security BasicAuth
// @Tags Recommendation
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recommendation/collaborative [get]
func (h *RecommendationHandler) GetRecommendationCollaborative(ctx *gin.Context) {
	userID := ctx.Value("userID").(string)
	recipes, err := h.recommendationUC.GetRecommendationCollaborative(ctx, userID)
	if err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteResponseWithBody(ctx, http.StatusOK, "success", recipes)
}

// @Summary Get recommendation by preferences
// @Description Get recommendation based on user preferences
// @Security BasicAuth
// @Tags Recommendation
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recommendation/preferences [get]
func (h *RecommendationHandler) GetRecommendationByPreferences(ctx *gin.Context) {
	userID := ctx.Value("userID").(string)
	recipes, err := h.recommendationUC.GetRecommendationPreferences(ctx, userID)
	if err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteResponseWithBody(ctx, http.StatusOK, "success", recipes)
}

type newInteractionRequest struct {
	RecipeID    string `json:"recipe_id" binding:"required"`
	Interaction *int   `json:"interaction" binding:"required"`
}

// @Summary Create new interaction
// @Description Create a new interaction for a recipe
// @Security BasicAuth
// @Tags Recommendation
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /recommendation/interaction/{userID} [post]
func (h *RecommendationHandler) NewInteraction(ctx *gin.Context) {
	var req newInteractionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID := ctx.Value("userID").(string)
	err := h.recommendationUC.NewInteraction(ctx, userID, req.RecipeID, *req.Interaction)
	if err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "success")
}
