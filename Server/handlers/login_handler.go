package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"Server/middleware"
	"Server/models"
	"Server/utils"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token  string         `json:"token"`
	Player *models.Player `json:"player"`
}

// HandleLogin 处理登录请求
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// 验证HTTP方法
	if r.Method != http.MethodPost {
		utils.SendError(w, r, 1010, "Method not allowed")
		return
	}

	// 验证必要的头部
	deviceID := r.Header.Get("Device-ID")
	reqID := r.Header.Get("Req-ID")

	if deviceID == "" || reqID == "" {
		utils.SendError(w, r, 1011, "Missing required headers")
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.SendError(w, r, 1013, "Failed to read request body")
		return
	}

	// 解析登录请求
	var loginReq LoginRequest
	if err := json.Unmarshal(body, &loginReq); err != nil {
		utils.SendError(w, r, 1014, "Invalid request format")
		return
	}

	// 获取玩家存储
	playerStore := models.GetPlayerStore()

	// 查找玩家
	player := playerStore.FindByUsername(loginReq.Username)

	// 如果玩家不存在，创建新玩家
	if player == nil {
		player = playerStore.Create(loginReq.Username, deviceID)
	}

	// 更新最后登录时间
	playerStore.UpdateLastLogin(player.ID)

	// 生成JWT令牌
	token, err := middleware.GenerateToken(player.ID)
	if err != nil {
		utils.SendError(w, r, 1016, "Failed to generate token")
		return
	}

	// 返回成功响应
	utils.SendSuccess(w, r, LoginResponse{
		Token:  token,
		Player: player,
	})
}
