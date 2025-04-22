package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SendSuccess 发送成功响应
func SendSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	// 设置响应头
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Req-ID", r.Header.Get("Req-ID"))
	w.Header().Set("Code", "0") // 成功代码

	// 序列化响应
	resp := Response{
		Code: 0,
		Data: data,
	}

	respBytes, _ := json.Marshal(resp)
	w.Write(respBytes)
}

// SendError 发送错误响应
func SendError(w http.ResponseWriter, r *http.Request, errorCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Req-ID", r.Header.Get("Req-ID"))
	w.Header().Set("Code", strconv.Itoa(errorCode))

	resp := Response{
		Code:    errorCode,
		Message: message,
	}

	respBytes, _ := json.Marshal(resp)
	w.Write(respBytes)
}
