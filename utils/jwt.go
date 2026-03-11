package utils

import (
	"bookadmin/global"
	"bookadmin/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   uint           `json:"user_id"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string, role model.UserRole) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(jwtDuration())

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    global.GVA_CONFIG.JWT.Issuer,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret())
	return token, err
}

// ParseToken 解析JWT token
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret(), nil
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == nil {
		return true
	}
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

func jwtSecret() []byte {
	if global.GVA_CONFIG == nil || global.GVA_CONFIG.JWT.Secret == "" {
		return []byte("bookadmin-secret-key-change-in-production")
	}
	return []byte(global.GVA_CONFIG.JWT.Secret)
}

func jwtDuration() time.Duration {
	if global.GVA_CONFIG == nil {
		return 24 * time.Hour
	}
	return global.GVA_CONFIG.JWTDuration()
}
