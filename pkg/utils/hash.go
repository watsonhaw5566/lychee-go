package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// MD5 计算 MD5 哈希
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA1 计算 SHA1 哈希
func SHA1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算 SHA256 哈希
func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// PasswordHash 使用 Bcrypt 加密密码
func PasswordHash(password string, cost ...int) (string, error) {
	c := bcrypt.DefaultCost
	if len(cost) > 0 && cost[0] >= bcrypt.MinCost && cost[0] <= bcrypt.MaxCost {
		c = cost[0]
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), c)
	return string(bytes), err
}

// PasswordVerify 验证密码
func PasswordVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SimpleHash 简单哈希（MD5 + Salt）
func SimpleHash(str string, salt ...string) string {
	s := ""
	if len(salt) > 0 {
		s = salt[0]
	}
	return MD5(str + s + "_lychee_go")
}
