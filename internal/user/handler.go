package user

import (
	"flove/job/internal/base/database"
	"flove/job/internal/base/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUC UserUC
}

func NewUserHandler(userUC UserUC) *UserHandler {
	return &UserHandler{
		userUC: userUC,
	}
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50" example:"John Shnow"`
	Email    string `json:"email" binding:"required,email" example:"example@gmail.com"`
	Password string `json:"password" binding:"required,min=6,max=32" example:"password"`
	Phone    string `json:"phone" binding:"required,e164"`
}

// @Summary Create a new user
// @Description Create a new user with the given information
// @Tags User
// @Accept json
// @Produce json
// @Param request body createUserRequest true "User information"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users [post]
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req createUserRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user := UserModel{
		Username:  req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user.SetPassword(req.Password)

	err = h.userUC.CreateUser(ctx, &user)
	if err != nil {
		response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, &response.Response{
		Code:    http.StatusCreated,
		Message: "user succesfully created",
	})
}

// @Summary Update user information
// @Description Update information of the currently logged in user
// @Security BasicAuthcAuth
// @Tags User
// @Accept json
// @Produce json
//
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users [patch]
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	var req struct {
		Phone *string `json:"phone" binding:"omitempty,e164"`
	}

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID := ctx.MustGet("userID").(string)

	err = h.userUC.UpdateUser(ctx, userID, UpdateUserDTO{
		Phone: req.Phone,
	})
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "user succesfully updated")
}

// @Summary Get user information
// @Description Get information about the currently logged in user
// @Security BasicAuth
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{userID} [get]
func (h *UserHandler) GetUserInfo(ctx *gin.Context) {
	userID := ctx.Param("userID")
	user, err := h.userUC.GetUserByID(ctx, userID)
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteResponseWithBody(ctx, http.StatusOK, "success", struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}{
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
	})
}

type changePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=32" example:"password"`
}

// @Summary Change user password
// @Description Change the password of the currently logged in user
// @Security BasicAuthcAuth
// @Tags User
// @Accept json
// @Produce json
//
//	@Param request body changePasswordRequest true "User information"
//
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/password [patch]
func (h *UserHandler) ChangePassword(ctx *gin.Context) {
	var req changePasswordRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID := ctx.MustGet("userID").(string)
	err = h.userUC.ChangePassword(ctx, userID, req.Password)
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "password succesfully changed")
}

type changeUserRoleRequest struct {
	UserID string `uri:"id" binding:"required" example:"21"`
	Role   *int   `json:"role" binding:"required" example:"0"`
}

// @Summary Change user role
// @Description Change the role of a user with the given ID
// @Security BasicAuth
// @Tags User
// @Accept json
// @Produce json
// @Param request body changeUserRoleRequest true "User information"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/users/role/{id} [patch]
func (h *UserHandler) ChangeUserRole(ctx *gin.Context) {
	var req changeUserRoleRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.WriteResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userUC.ChangeUserRole(ctx, req.UserID, Role(*req.Role)); err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "role successfully changed")
}

// @Summary Delete user
// @Description Delete the currently logged in user
// @Security BasicAuth
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users [delete]
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	err := h.userUC.DeleteUser(ctx, userID)
	if err != nil {
		switch err {
		case database.ErrNotFound:
			response.WriteResponse(ctx, http.StatusNotFound, err.Error())
		default:
			response.WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteResponse(ctx, http.StatusOK, "user succesfully deleted")
}
