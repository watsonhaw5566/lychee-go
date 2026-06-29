package service

import (
	"lychee-go/app/model"
	"lychee-go/internal/logger"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetList(page int, pageSize int, name string) ([]model.User, int64, error) {
	logger.Debug("Query user list: page=%d, pageSize=%d, name=%s", page, pageSize, name)
	return model.GetUserList(page, pageSize, name)
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return model.GetUserByID(id)
}

func (s *UserService) Create(user *model.User) error {
	logger.Info("Creating user: name=%s, email=%s", user.Name, user.Email)
	return model.CreateUser(user)
}

func (s *UserService) Update(id uint, updates map[string]interface{}) error {
	logger.Info("Updating user: id=%d, updates=%v", id, updates)
	return model.UpdateUser(id, updates)
}

func (s *UserService) Delete(id uint) error {
	logger.Info("Deleting user: id=%d", id)
	return model.DeleteUser(id)
}

func (s *UserService) GetByEmail(email string) (*model.User, error) {
	return model.GetUserByEmail(email)
}
