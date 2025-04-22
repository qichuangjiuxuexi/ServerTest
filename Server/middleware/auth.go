package middleware

// Claims 定义JWT中的声明
type Claims struct {
	PlayerID string `json:"player_id"`
}

// GenerateToken 为玩家生成JWT令牌
func GenerateToken(playerID string) (string, error) {

	return playerID, nil
}
