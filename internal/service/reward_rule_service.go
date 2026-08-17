package service

import (
	"errors"
	"sort"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/internal/store"
	"checkinkeeper/pkg/idgen"
)

// CreateRewardRule 创建奖励规则；若指定了 ActivityID 则校验活动存在。
func (s *Service) CreateRewardRule(input model.RewardRule) (*model.RewardRule, error) {
	input.ID = ""
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if input.ActivityID != "" {
		if _, err := s.store.GetActivity(input.ActivityID); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return nil, model.NewValidationError("activity_id", "关联的活动不存在")
			}
			return nil, err
		}
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateRewardRule(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建奖励规则 %s(%s)", input.Name, input.Type)
	return &input, nil
}

// GetRewardRule 查询奖励规则详情。
func (s *Service) GetRewardRule(id string) (*model.RewardRule, error) {
	return s.store.GetRewardRule(id)
}

// ListRewardRules 分页查询奖励规则列表。
func (s *Service) ListRewardRules(filter model.RewardRuleFilter, page, size int) ([]*model.RewardRule, int, error) {
	all := s.store.ListRewardRules()
	matched := make([]*model.RewardRule, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.RewardRule{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateRewardRule 更新奖励规则字段。
func (s *Service) UpdateRewardRule(id, name, status string, threshold int, points int64) (*model.RewardRule, error) {
	r, err := s.store.GetRewardRule(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		r.Name = name
	}
	if status != "" {
		r.Status = status
	}
	if threshold > 0 {
		r.Threshold = threshold
	}
	if points > 0 {
		r.Points = points
	}
	r.UpdatedAt = time.Now()
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRewardRule(r); err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteRewardRule 删除奖励规则。
func (s *Service) DeleteRewardRule(id string) error {
	if _, err := s.store.GetRewardRule(id); err != nil {
		return err
	}
	return s.store.DeleteRewardRule(id)
}
