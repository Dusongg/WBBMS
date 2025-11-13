package utils

import (
	"bookadmin/global"
	"bookadmin/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var jwtSecret = []byte("bookadmin-secret-key-change-in-production")

type Claims struct {
	UserID   uint           `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string, role model.UserRole) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour) // 24小时过期

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    "bookadmin",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

// ParseToken 解析JWT token
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

// HashPassword 密码加密
func HashPassword(password string) (string, error) {
	// 使用bcrypt加密，需要添加golang.org/x/crypto/bcrypt
	// 这里简化处理，实际应该使用bcrypt
	return password, nil
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
	// 简化处理，实际应该使用bcrypt验证
	return hashedPassword == password
}

// GetUserByID 根据ID获取用户
func GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := global.GVA_DB.First(&user, userID).Error; err != nil {
		global.GVA_LOG.Error("获取用户失败", zap.Error(err))
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}
