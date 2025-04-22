package middleware

import (
	"time"

	"github.com/golang-jwt/jwt/v4"

	"Server/config"
)

// Claims 定义JWT中的声明
type Claims struct {
	PlayerID string `json:"player_id"`
	jwt.RegisteredClaims
}

// GenerateToken 为玩家生成JWT令牌
func GenerateToken(playerID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		PlayerID: playerID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetConfig().JWTKey))
}
