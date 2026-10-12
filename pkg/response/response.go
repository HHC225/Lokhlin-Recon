// Package response -----------------------------
// @file      : response.go
// @time      : 2025/5/4 22:04
// -------------------------------------------
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
	})
}

func OkWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
	})
}

func OkWithData(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func OkWithMessageAndData(message string, data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

func Fail(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    500,
		Message: "fail",
	})
}

func FailWithMessage(message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    500,
		Message: message,
	})
}

func FailWithCodeAndMessage(code int, message string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func FailWithCodeAndMessageAndData(code int, message string, data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
