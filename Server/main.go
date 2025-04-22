package main

import (
	"fmt"
	"log"
	"net/http"

	"Server/config"
	"Server/handlers"
)

func main() {
	// 加载配置
	cfg := config.GetConfig()

	// 创建HTTP服务多路复用器
	mux := http.NewServeMux()

	// 注册路由 - 只包含登录接口
	mux.HandleFunc("/player/login", handlers.HandleLogin)

	// 启动服务器
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("Server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
