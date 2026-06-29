package controller

import (
	"strconv"

	"lychee-go/app/model"
	"lychee-go/app/service"
	"lychee-go/internal/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

// GetList 用户列表 GET /api/users
func (ctrl *UserController) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	name := c.Query("name")

	users, total, err := ctrl.userService.GetList(page, pageSize, name)
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}

	response.Paginate(c, users, total, page, pageSize)
}

// GetUser 获取单个用户 GET /api/users/:id
func (ctrl *UserController) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	user, err := ctrl.userService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.SuccessWithData(c, user)
}

// CreateUser 创建用户 POST /api/users
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Status   int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Status:   req.Status,
	}

	if err := ctrl.userService.Create(user); err != nil {
		response.ServerError(c, "创建失败: "+err.Error())
		return
	}

	response.SuccessWithData(c, user)
}

// UpdateUser 更新用户 PUT /api/users/:id
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	var req struct {
		Name   string `json:"name"`
		Email  string `json:"email" binding:"omitempty,email"`
		Status *int   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := ctrl.userService.Update(uint(id), updates); err != nil {
		response.ServerError(c, "更新失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功")
}

// DeleteUser 删除用户 DELETE /api/users/:id
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	if err := ctrl.userService.Delete(uint(id)); err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功")
}
