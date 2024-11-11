package response

import "github.com/gin-gonic/gin"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    any    `json:"body,omitempty"`
}

const (
	BadRequestMessage          = "Bad request"
	InternalServerErrorMessage = "Internal server error"
	NotFoundMessage            = "Resource not found"
)

func WriteResponse(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, &Response{
		Code:    status,
		Message: message,
	})
}

func WriteResponseWithBody(ctx *gin.Context, status int, message string, body any) {
	ctx.JSON(status, &Response{
		Code:    status,
		Message: message,
		Body:    body,
	})
}
