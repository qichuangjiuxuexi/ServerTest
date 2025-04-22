package models

import (
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

// GetPlayerStore 获取玩家存储的单例实例
func GetPlayerStore() *PlayerStore {
	once.Do(func() {
		var err error

		parentDir := tools.GetSqlPath()

		// 在父目录中设置Sql文件夹路径
		sqlDirPath := filepath.Join(parentDir, "Sql")
		playerListPath := filepath.Join(sqlDirPath, "PlayerList.txt")

		// 确保Sql文件夹存在
		if err = tools.EnsureDirectoryExists(sqlDirPath); err != nil {
			panic(fmt.Errorf("确保SQL目录存在失败: %w", err))
		}

		// 确保PlayerList.txt文件存在
		if err = tools.EnsureFileExists(playerListPath); err != nil {
			panic(fmt.Errorf("确保PlayerList.txt文件存在失败: %w", err))
		}

		fmt.Printf("初始化PlayerStore成功，SQL目录：%s，玩家列表文件：%s\n", sqlDirPath, playerListPath)

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

	// 检查文件是否存在，不存在则创建
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

// Create 创建新玩家
func (s *PlayerStore) Create(username, deviceID string) *Player {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 首先检查玩家是否已存在
	for _, player := range s.players {
		if player.DeviceID == deviceID {
			fmt.Printf("设备ID %s 的玩家已存在，ID: %s\n", deviceID, player.ID)
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

	id := deviceID
	now := time.Now()

	// 打开 PlayerList.txt 文件（如果不存在则创建）
	file, err := os.OpenFile(s.playerListPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("打开 PlayerList.txt 文件失败: %v\n", err)
		return nil
	}
	defer file.Close()

	// 写入玩家信息前先检查文件中是否已存在该玩家
	// 重新打开文件用于读取
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
		parts := strings.Split(line, ",")
		if len(parts) >= 2 && (parts[0] == id || parts[1] == deviceID) {
			playerExists = true
			break
		}
	}
	checkFile.Close()

	if !playerExists {
		// 写入玩家信息
		_, err = file.WriteString(fmt.Sprintf("%s,%s\n", id, deviceID))
		if err != nil {
			fmt.Printf("写入玩家信息失败: %v\n", err)
			return nil
		}
		fmt.Printf("新玩家信息已写入文件: ID=%s, DeviceID=%s\n", id, deviceID)
	} else {
		fmt.Println("玩家在文件中已存在，跳过写入")
	}

	// 创建或获取玩家对象
	player, exists := s.players[id]
	if !exists {
		player = &Player{
			ID:          id,
			Username:    username,
			DeviceID:    deviceID,
			CreatedAt:   now,
			LastLoginAt: now,
		}
		s.players[id] = player
		fmt.Printf("创建新玩家: ID=%s, Username=%s\n", id, username)
	} else {
		player.LastLoginAt = now
		fmt.Printf("更新现有玩家登录时间: ID=%s\n", id)
	}

	return player
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
