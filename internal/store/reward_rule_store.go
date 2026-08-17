package store

import (
	"checkinkeeper/internal/model"
)

func (s *MemoryStore) CreateRewardRule(r *model.RewardRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rewardRules[r.ID] = r
	return nil
}

func (s *MemoryStore) GetRewardRule(id string) (*model.RewardRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rewardRules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListRewardRules() []*model.RewardRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.RewardRule, 0, len(s.rewardRules))
	for _, r := range s.rewardRules {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) UpdateRewardRule(r *model.RewardRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rewardRules[r.ID]; !ok {
		return ErrNotFound
	}
	s.rewardRules[r.ID] = r
	return nil
}

func (s *MemoryStore) DeleteRewardRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rewardRules[id]; !ok {
		return ErrNotFound
	}
	delete(s.rewardRules, id)
	return nil
}
