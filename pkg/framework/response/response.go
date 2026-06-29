package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginateResponse 分页响应格式
type PaginateResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    PaginateData `json:"data"`
}

// PaginateData 分页数据结构
type PaginateData struct {
	List      interface{} `json:"list"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	TotalPage int64       `json:"total_page"`
}

// ======== 成功响应 ========

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success"})
}

func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func SuccessWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: message})
}

// ======== 错误响应 ========

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{Code: code, Message: message})
}

func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: code, Message: message, Data: data})
}

// ======== HTTP 状态码响应 ========

func NotFound(c *gin.Context, message ...string) {
	msg := "资源不存在"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusNotFound, Response{Code: 404, Message: msg})
}

func Forbidden(c *gin.Context, message ...string) {
	msg := "无权限访问"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusForbidden, Response{Code: 403, Message: msg})
}

func Unauthorized(c *gin.Context, message ...string) {
	msg := "未登录"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusUnauthorized, Response{Code: 401, Message: msg})
}

func BadRequest(c *gin.Context, message ...string) {
	msg := "参数错误"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusBadRequest, Response{Code: 400, Message: msg})
}

func ServerError(c *gin.Context, message ...string) {
	msg := "服务器内部错误"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: msg})
}

// ======== 分页响应 ========

func Paginate(c *gin.Context, list interface{}, total int64, page int, pageSize int) {
	totalPage := int64(0)
	if pageSize > 0 {
		totalPage = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	c.JSON(http.StatusOK, PaginateResponse{
		Code:    0,
		Message: "success",
		Data: PaginateData{
			List:      list,
			Total:     total,
			Page:      page,
			PageSize:  pageSize,
			TotalPage: totalPage,
		},
	})
}

// ======== 自定义响应 ========

func JSON(c *gin.Context, httpCode int, code int, message string, data interface{}) {
	c.JSON(httpCode, Response{Code: code, Message: message, Data: data})
}

func Raw(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
