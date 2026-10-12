package response

import (
	"fmt"
	"net/http"
	"recon/internal/logger"
	"runtime"

	"github.com/gin-gonic/gin"
	"recon/internal/i18n"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SuccessResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"Operation succeeded"`
	Data    interface{} `json:"data,omitempty"`
}

type BadRequestResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request"`
}

type UnauthorizedResponse struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"Unauthorized access"`
}

type NotFoundResponse struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Resource not found"`
}

type InternalServerErrorResponse struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal server error"`
}

type PageResponse struct {
	Total    int64       `json:"total"`
	List     interface{} `json:"list"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func getLocale(c *gin.Context) string {
	if locale := c.GetString("language"); locale != "" {
		return locale
	}
	locale := c.GetHeader("Accept-Language")
	if locale == "" {
		locale = "ko-KR"
	}
	return locale
}

func Success(c *gin.Context, data interface{}, msgKey string) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}, msgKey string) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    data,
	})
}

func BadRequest(c *gin.Context, msgKey string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    errMsg,
	})
}

func NotFound(c *gin.Context, msgKey string, err error) {
	c.JSON(http.StatusNotFound, Response{
		Code:    http.StatusNotFound,
		Message: i18n.Translate(getLocale(c), msgKey),
	})
}

func InternalServerError(c *gin.Context, msgKey string, err error) {
	pc, file, line, ok := runtime.Caller(1)
	callerInfo := ""
	if ok {
		fn := runtime.FuncForPC(pc)
		callerInfo = fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
	}

	logger.Error(fmt.Sprintf("InternalServerError at %s: %v", callerInfo, err))
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    fmt.Sprintf("%v", err),
	})
}

func Unauthorized(c *gin.Context, msgKey string, err error) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    http.StatusUnauthorized,
		Message: i18n.Translate(getLocale(c), msgKey),
		Data:    fmt.Sprintf("%v", err),
	})
}
