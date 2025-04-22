package models

import (
	"strconv"
	"time"
)

// Player 表示玩家模型
type Player struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DeviceID    string    `json:"device_id"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// 生成唯一ID (在实际应用中可能使用UUID库)
func generateID() string {
	return time.Now().Format("20060102150405") +
		strconv.FormatInt(time.Now().UnixNano()%1000, 10)
}
