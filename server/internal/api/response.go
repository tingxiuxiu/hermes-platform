package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    201,
		Message: "created",
		Data:    data,
	})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Code:    400,
		Message: "bad request",
		Error:   err,
	})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Code:    401,
		Message: "unauthorized",
		Error:   err,
	})
}

func Forbidden(c *gin.Context, err string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Code:    403,
		Message: "forbidden",
		Error:   err,
	})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Code:    404,
		Message: "not found",
		Error:   err,
	})
}

func InternalServerError(c *gin.Context, err string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Code:    500,
		Message: "internal server error",
		Error:   err,
	})
}

func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}
