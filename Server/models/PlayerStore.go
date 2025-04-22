package models

import (
	"sync"
	"time"
)

// PlayerStore 管理玩家数据
type PlayerStore struct {
	mutex   sync.RWMutex
	players map[string]*Player
}

var store *PlayerStore
var once sync.Once

// GetPlayerStore 获取玩家存储的单例实例
func GetPlayerStore() *PlayerStore {
	once.Do(func() {
		store = &PlayerStore{
			players: make(map[string]*Player),
		}
	})
	return store
}

// FindByUsername 通过用户名查找玩家
func (s *PlayerStore) FindByUserId(userId string) *Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, p := range s.players {
		if p.ID == userId {
			return p
		}
	}
	return nil
}

// Create 创建新玩家
func (s *PlayerStore) Create(username, deviceID string) *Player {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := deviceID
	now := time.Now()

	player := &Player{
		ID:          id,
		Username:    username,
		DeviceID:    deviceID,
		CreatedAt:   now,
		LastLoginAt: now,
	}

	s.players[id] = player
	return player
}

// UpdateLastLogin 更新玩家最后登录时间
func (s *PlayerStore) UpdateLastLogin(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if player, ok := s.players[id]; ok {
		player.LastLoginAt = time.Now()
		return true
	}
	return false
}
