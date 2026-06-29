package model

import (
	"lychee-go/internal/db"
)

type User struct {
	db.BaseModel
	Name     string `gorm:"column:name;size:100;not null" json:"name"`
	Email    string `gorm:"column:email;size:100;uniqueIndex" json:"email"`
	Password string `gorm:"column:password;size:255" json:"-"`
	Status   int    `gorm:"column:status;default:1" json:"status"`
}

func (User) TableName() string {
	return "users"
}

// ======== 查询方法 ========

func GetUserByID(id uint) (*User, error) {
	var user User
	err := db.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(user *User) error {
	return db.DB.Create(user).Error
}

func UpdateUser(id uint, updates map[string]interface{}) error {
	return db.DB.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteUser(id uint) error {
	return db.DB.Delete(&User{}, id).Error
}

func GetUserList(page int, pageSize int, name string) ([]User, int64, error) {
	var users []User
	var total int64

	query := db.DB.Model(&User{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error

	return users, total, err
}

func AutoMigrate() error {
	return db.DB.AutoMigrate(&User{})
}
