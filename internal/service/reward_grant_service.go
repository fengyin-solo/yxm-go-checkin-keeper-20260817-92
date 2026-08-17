package service

import (
	"sort"

	"checkinkeeper/internal/model"
)

// GetRewardGrant 查询奖励发放流水详情。
func (s *Service) GetRewardGrant(id string) (*model.RewardGrant, error) {
	return s.store.GetRewardGrant(id)
}

// ListRewardGrants 分页查询奖励发放流水，按时间倒序。
func (s *Service) ListRewardGrants(filter model.GrantFilter, page, size int) ([]*model.RewardGrant, int, error) {
	all := s.store.ListRewardGrants()
	matched := make([]*model.RewardGrant, 0, len(all))
	for _, g := range all {
		if filter.Match(g) {
			matched = append(matched, g)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.Before(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.RewardGrant{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
