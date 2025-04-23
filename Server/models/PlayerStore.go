package models

import (
	"encoding/json"
	"Server/tools"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PlayerStore 管理玩家数据
type PlayerStore struct {
	mutex          sync.RWMutex
	players        map[string]*Player
	sqlDirPath     string
	playerListPath string
}

var store *PlayerStore
var once sync.Once

// GetPlayerStore 获取玩家存储的单例实�?
func GetPlayerStore() *PlayerStore {
	once.Do(func() {
		var err error

		parentDir := tools.GetSqlPath()

		// 在父目录中设置Sql文件夹路�?
		sqlDirPath := filepath.Join(parentDir, "Sql")
		playerListPath := filepath.Join(sqlDirPath, "PlayerList.txt")

		// 确保Sql文件夹存�?
		if err = tools.EnsureDirectoryExists(sqlDirPath); err != nil {
			panic(fmt.Errorf("确保SQL目录存在失败: %w", err))
		}

		// 确保PlayerList.txt文件存在
		if err = tools.EnsureFileExists(playerListPath); err != nil {
			panic(fmt.Errorf("确保PlayerList.txt文件存在失败: %w", err))
		}

		fmt.Printf("初始化PlayerStore成功，SQL目录�?%s，玩家列表文件：%s\n", sqlDirPath, playerListPath)

		store = &PlayerStore{
			players:        make(map[string]*Player),
			sqlDirPath:     sqlDirPath,
			playerListPath: playerListPath,
		}
	})
	return store
}

// FindByUserId 通过用户ID找玩家
func (s *PlayerStore) FindByUserId(userId string) *Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if err := tools.EnsureFileExists(s.playerListPath); err != nil {
		fmt.Printf("查找用户前检查文件失败: %v\n", err)
		return nil
	}

	// 打开 PlayerList.txt 文件
	file, err := os.OpenFile(s.playerListPath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("打开 PlayerList.txt 文件失败: %v\n", err)
		return nil
	}
	defer file.Close()

	// 读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // 跳过空行
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 2 { // 使用>=2可以兼容将来可能添加更多字段
			id := parts[0]
			deviceID := parts[1]
			if id == userId || deviceID == userId {
				// 如果在文件中找到但内存中不存在，则加载到内存
				if _, exists := s.players[id]; !exists {
					// 基于文件记录创建Player对象
					s.mutex.RUnlock() // 解除读锁
					s.mutex.Lock()    // 获取写锁
					// 再次检查，避免锁切换期间其他goroutine已经创建
					if _, exists := s.players[id]; !exists {
						s.players[id] = &Player{
							ID:        id,
							DeviceID:  deviceID,
							Username:  "Player_" + id, // 使用ID前8位作为默认用户名
							CreatedAt: time.Now(),     // 使用当前时间作为默认创建时间
						}
					}
					player := s.players[id]
					s.mutex.Unlock() // 释放写锁
					s.mutex.RLock()  // 重新获取读锁
					return player
				}
				return s.players[id]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取PlayerList.txt文件时发生错误: %v\n", err)
	}

	return nil
}

func (s *PlayerStore) GetPlayerCount() int {
	if err := tools.EnsureFileExists(s.playerListPath); err != nil {
		fmt.Printf("查找用户前检查文件失败: %v\n", err)
		return 0
	}

	// 打开 PlayerList.txt 文件
	file, err := os.OpenFile(s.playerListPath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("打开 PlayerList.txt 文件失败: %v\n", err)
		return 0
	}
	defer file.Close()

	// 读取文件内容
	scanner := bufio.NewScanner(file)
	var count = 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break // 跳过空行
		} else {
			count++
		}
	}
	return count
}

// Create 创建新玩家
func (s *PlayerStore) Create(username, deviceID string) *Player {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 首先检查玩家是否已存在
	for _, player := range s.players {
		if player.DeviceID == deviceID {
			// 更新现有玩家的登录时间
			player.LastLoginAt = time.Now()
			fmt.Printf("设备ID %s 的玩家已存在，ID: %s, 更新登录时间\n", deviceID, player.ID)

			// 更新文件中的 lastLoginAt
			if err := s.updatePlayerInFile(player.ID, player.LastLoginAt); err != nil {
				fmt.Printf("更新文件中的玩家数据失败: %v\n", err)
				return nil
			}

			// 返回玩家信息，包含 CreatedAt 和 LastLoginAt
			return player
		}
	}

	// 检查目录和文件是否存在
	if err := tools.EnsureDirectoryExists(s.sqlDirPath); err != nil {
		fmt.Printf("创建玩家前确保目录存在失败: %v\n", err)
		return nil
	}

	if err := tools.EnsureFileExists(s.playerListPath); err != nil {
		fmt.Printf("创建玩家前确保文件存在失败: %v\n", err)
		return nil
	}

	var id = fmt.Sprintf("%d", 100000+s.GetPlayerCount())
	now := time.Now()

	// 打开 PlayerList.txt 文件（如果不存在则创建）
	file, err := os.OpenFile(s.playerListPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("打开 PlayerList.txt 文件失败: %v\n", err)
		return nil
	}
	defer file.Close()

	// 重新打开文件用于读取并检查玩家是否存在
	checkFile, err := os.OpenFile(s.playerListPath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("检查玩家是否存在时打开文件失败: %v\n", err)
		return nil
	}

	scanner := bufio.NewScanner(checkFile)
	playerExists := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// 解析每行JSON格式的数据
		var playerInfo map[string]string
		if err := json.Unmarshal([]byte(line), &playerInfo); err != nil {
			fmt.Printf("解析玩家信息失败: %v\n", err)
			continue
		}

		// 检查是否已经存在相同的ID或deviceID
		if playerInfo["user"] == id || playerInfo["password"] == deviceID {
			playerExists = true
			break
		}
	}
	checkFile.Close()

	if !playerExists {
		// 写入新的玩家信息为JSON格式
		playerInfo := map[string]string{
			"user":        id,
			"password":    deviceID,
			"createdAt":   now.Format(time.RFC3339), // 格式化为ISO8601格式
			"lastLoginAt": now.Format(time.RFC3339),
		}

		playerInfoJSON, err := json.Marshal(playerInfo)
		if err != nil {
			fmt.Printf("将玩家信息转换为JSON失败: %v\n", err)
			return nil
		}

		_, err = file.WriteString(string(playerInfoJSON) + "\n")
		if err != nil {
			fmt.Printf("写入玩家信息失败: %v\n", err)
			return nil
		}
		fmt.Printf("新玩家信息已写入文件: ID=%s, DeviceID=%s\n", id, deviceID)
	} else {
		fmt.Println("玩家在文件中已存在，跳过写入")
	}

	// 获取玩家对象
	player, exists := s.players[id]
	if !exists {
		// 创建玩家对象用于返回数据
		player = &Player{
			ID:          id,
			Username:    "Player" + id,
			DeviceID:    deviceID,
			CreatedAt:   now,
			LastLoginAt: now,
		}
		s.players[id] = player
		fmt.Printf("创建新玩家: ID=%s, Username=%s\n", id, username)
	} else {
		// 更新现有玩家的登录时间
		player.LastLoginAt = now
		fmt.Printf("更新现有玩家登录时间: ID=%s\n", id)
	}

	// 返回玩家信息，包括 CreatedAt 和 LastLoginAt
	return player
}

// 辅助方法：更新指定玩家的 lastLoginAt 字段
func (s *PlayerStore) updatePlayerInFile(playerID string, lastLoginAt time.Time) error {
	// 创建临时文件
	tempFilePath := s.playerListPath + ".temp"
	tempFile, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开临时文件失败: %v", err)
	}
	defer tempFile.Close()

	// 打开原文件读取内容
	file, err := os.OpenFile(s.playerListPath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开原文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析每行JSON格式的数据
		var playerInfo map[string]string
		if err := json.Unmarshal([]byte(line), &playerInfo); err != nil {
			return fmt.Errorf("解析玩家信息失败: %v", err)
		}

		// 如果玩家ID匹配，更新 lastLoginAt
		if playerInfo["user"] == playerID {
			playerInfo["lastLoginAt"] = lastLoginAt.Format(time.RFC3339)
		}

		// 写入修改后的玩家信息到临时文件
		playerInfoJSON, err := json.Marshal(playerInfo)
		if err != nil {
			return fmt.Errorf("将玩家信息转换为JSON失败: %v", err)
		}

		_, err = tempFile.WriteString(string(playerInfoJSON) + "\n")
		if err != nil {
			return fmt.Errorf("写入临时文件失败: %v", err)
		}
	}

	// 检查文件是否读取完毕
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("扫描文件错误: %v", err)
	}

	// 替换原文件为临时文件
	if err := os.Rename(tempFilePath, s.playerListPath); err != nil {
		return fmt.Errorf("替换文件失败: %v", err)
	}

	return nil
}




// UpdateLastLogin 更新玩家最后登录时间
func (s *PlayerStore) UpdateLastLogin(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if player, ok := s.players[id]; ok {
		player.LastLoginAt = time.Now()
		fmt.Printf("更新玩家 %s 的最后登录时间\n", id)
		return true
	}

	fmt.Printf("未找到玩家 %s，无法更新登录时间\n", id)
	return false
}
