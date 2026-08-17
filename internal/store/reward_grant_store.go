package store

import (
	"checkinkeeper/internal/model"
)

func (s *MemoryStore) CreateRewardGrant(g *model.RewardGrant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rewardGrants[g.ID] = g
	return nil
}

func (s *MemoryStore) GetRewardGrant(id string) (*model.RewardGrant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.rewardGrants[id]
	if !ok {
		return nil, ErrNotFound
	}
	return g, nil
}

func (s *MemoryStore) ListRewardGrants() []*model.RewardGrant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.RewardGrant, 0, len(s.rewardGrants))
	for _, g := range s.rewardGrants {
		list = append(list, g)
	}
	return list
}
