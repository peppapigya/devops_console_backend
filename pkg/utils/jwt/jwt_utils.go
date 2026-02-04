package jwt

import (
	"devops-console-backend/internal/common"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// jwt 工具类

type Claims struct {
	ID       int64
	Username string
	Roles    []string
	jwt.RegisteredClaims
}

// GetUserId 获取当前登录用户id
func (claims *Claims) GetUserId() int64 {
	if claims == nil {
		return 0
	}
	return claims.ID
}

// GetUserName 获取当前登录用户名
func (claims *Claims) GetUserName() string {
	if claims == nil {
		return ""
	}
	return claims.Username
}

func (claims *Claims) GetRoles() []string {
	if claims == nil {
		return nil
	}
	return claims.Roles
}

// GenerateJwtToken 生成jwt token
func GenerateJwtToken(ID int64, username string, roles []string) (string, error) {
	jwtProperties := common.GetGlobalConfig().Jwt
	claims := Claims{
		ID:       ID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(jwtProperties.ExpireTime))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "k8s-platform-go",
		},
	}
	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtProperties.Secret))
}

// ParseToken 解析token
func ParseToken(tokenStr string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(common.GetGlobalConfig().Jwt.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		log.Printf("token 无效")
		return nil, jwt.ErrTokenNotValidYet
	}
	return &claims, nil
}

// GenerateRefreshToken 生成refresh token
func GenerateRefreshToken(ID int64) (string, error) {
	jwtProperties := common.GetGlobalConfig().Jwt
	claims := Claims{
		ID: ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(jwtProperties.RefreshExpireTime))), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                                   // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                                   // 签发时间
			Issuer:    "k8s-platform-go",                                                                                // 签发人，也就是区分
			Subject:   fmt.Sprintf("refresh_%d", ID),                                                                    //标识该token是刷新token
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtProperties.Secret))
}

// RefreshToken 刷新token
func (claims *Claims) RefreshToken() (string, error) {
	// 判断刷新 Token 是否过期
	if time.Now().After(claims.ExpiresAt.Time) {
		return "", jwt.ErrTokenNotValidYet
	}
	newAccessToken, err := GenerateJwtToken(claims.ID, claims.Username, claims.Roles)
	if err != nil {
		return "", err
	}
	return newAccessToken, err
}
