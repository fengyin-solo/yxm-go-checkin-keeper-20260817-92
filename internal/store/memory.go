package store

import (
	"sync"

	"checkinkeeper/internal/model"
)

// MemoryStore 基于内存 map 的 Store 实现，所有操作线程安全。
type MemoryStore struct {
	mu           sync.RWMutex
	users        map[string]*model.User
	activities   map[string]*model.Activity
	checkins     map[string]*model.CheckinRecord
	rewardRules  map[string]*model.RewardRule
	rewardGrants map[string]*model.RewardGrant
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[string]*model.User),
		activities:   make(map[string]*model.Activity),
		checkins:     make(map[string]*model.CheckinRecord),
		rewardRules:  make(map[string]*model.RewardRule),
		rewardGrants: make(map[string]*model.RewardGrant),
	}
}

var _ Store = (*MemoryStore)(nil)
