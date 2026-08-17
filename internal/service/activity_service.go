package service

import (
	"sort"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/internal/store"
	"checkinkeeper/pkg/idgen"
)

// CreateActivity 创建签到活动（初始为草稿态）。
func (s *Service) CreateActivity(input model.Activity) (*model.Activity, error) {
	input.ID = ""
	input.Status = model.ActivityDraft
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateActivity(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建签到活动 %s(%s)", input.Name, input.Code)
	return &input, nil
}

// GetActivity 查询活动详情。
func (s *Service) GetActivity(id string) (*model.Activity, error) {
	return s.store.GetActivity(id)
}

// ListActivities 分页查询活动列表。
func (s *Service) ListActivities(filter model.ActivityFilter, page, size int) ([]*model.Activity, int, error) {
	all := s.store.ListActivities()
	matched := make([]*model.Activity, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Activity{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateActivity 更新活动可编辑字段（已结束活动不可编辑）。
func (s *Service) UpdateActivity(id, name, description string) (*model.Activity, error) {
	a, err := s.store.GetActivity(id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.ActivityFinished {
		return nil, model.NewValidationError("status", "已结束活动不可编辑")
	}
	if name != "" {
		a.Name = name
	}
	if description != "" {
		a.Description = description
	}
	a.UpdatedAt = time.Now()
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}

// TransitionActivity 活动状态流转（draft->active->finished）。
func (s *Service) TransitionActivity(id, to string) (*model.Activity, error) {
	a, err := s.store.GetActivity(id)
	if err != nil {
		return nil, err
	}
	if !model.CanActivityTransition(a.Status, to) {
		return nil, model.NewValidationError("status", "活动状态不允许从 "+a.Status+" 流转到 "+to)
	}
	a.Status = to
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateActivity(a); err != nil {
		return nil, err
	}
	s.log.Infof("活动 %s 状态流转为 %s", a.Code, to)
	return a, nil
}

// DeleteActivity 删除活动；若已有签到记录则拒绝删除。
func (s *Service) DeleteActivity(id string) error {
	if _, err := s.store.GetActivity(id); err != nil {
		return err
	}
	for _, c := range s.store.ListCheckinRecords() {
		if c.ActivityID == id {
			return store.ErrConflict
		}
	}
	return s.store.DeleteActivity(id)
}
