package auth

import (
	"errors"
	"flove/job/internal/base/database"
	"flove/job/internal/base/response"
	"flove/job/internal/user"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tokenUC TokenUC
	userUC  user.UserUC
}

func NewTokenHandler(tokenUC TokenUC, userUC user.UserUC) *AuthHandler {
	return &AuthHandler{
		tokenUC: tokenUC,
		userUC:  userUC,
	}
}

type signInRequest struct {
	Email    string `json:"email" binding:"required" example:"example@gmail.com"`
	Password string `json:"password" binding:"required" example:"password"`
}

// @Summary Sign in
// @Description Sign in with credentials and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body signInRequest true "Credentials"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 401 {string} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/signin [post]
func (h *AuthHandler) SignIn(ctx *gin.Context) {
	var req signInRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	refreshToken, err := h.tokenUC.NewRefreshToken(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrMismatchedPassword) || errors.Is(err, database.ErrNotFound):
			response.WriteResponse(ctx, http.StatusUnauthorized, err.Error())
			return
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}

	accessToken, err := h.tokenUC.NewAccessToken(ctx, refreshToken.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &response.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	ctx.SetCookie("access_token", string(accessToken.Token), int(time.Hour.Seconds()), "/", "", false, true)
	ctx.SetCookie("refresh_token", string(refreshToken.Token), int(time.Hour.Seconds()*24), "/", "", false, true)

	response.WriteResponse(ctx, http.StatusOK, "token succesfully created")
}

// @Summary Sign out
// @Description Sign out and delete tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/signout [post]
func (h *AuthHandler) SignOut(ctx *gin.Context) {
	refreshToken, _ := ctx.Cookie("refresh_token")

	err := h.tokenUC.DeleteRefreshToken(ctx, refreshToken)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		ctx.JSON(http.StatusInternalServerError, &response.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	ctx.SetCookie("access_token", "", -1, "/", "", false, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, &response.Response{
		Code:    http.StatusOK,
		Message: "token succesfully deleted",
	})
}

func (h *AuthHandler) DeleteTokens(ctx *gin.Context) {
	panic("unimplemented")
}

func (h *AuthHandler) RequireAuthenticatedUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ok := h.authenticate(ctx)
		if !ok {
			return
		}

		ctx.Next()
	}
}

func (h *AuthHandler) RequireRole(requiredRole user.Role) gin.HandlerFunc {
	h.RequireAuthenticatedUser()
	return func(ctx *gin.Context) {
		ok := h.authenticate(ctx)
		if !ok {
			return
		}

		value, _ := ctx.Get("role")

		role := value.(user.Role)
		if role < requiredRole {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Code:    http.StatusUnauthorized,
				Message: "not enought rights",
			})
			return
		}

		ctx.Next()
	}
}

func (h *AuthHandler) authenticate(ctx *gin.Context) bool {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
			Code:    http.StatusUnauthorized,
			Message: "refresh token is required",
		})
		return false
	}

	accessToken, _ := ctx.Cookie("access_token")
	userUUID, role, err := h.tokenUC.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			token, err := h.tokenUC.NewAccessToken(ctx, refreshToken)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
					Code:    http.StatusUnauthorized,
					Message: "refresh token is invalid",
				})
				return false
			}

			ctx.SetCookie("access_token", string(token.Token), int(time.Hour.Seconds()), "/", "", false, true)
			userUUID = token.UserUUID
			role = token.Role
		default:
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return false
		}
	}

	ctx.Set("userID", userUUID)
	ctx.Set("role", role)

	return true
}
