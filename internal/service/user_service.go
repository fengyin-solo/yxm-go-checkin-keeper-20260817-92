package service

import (
	"sort"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/internal/store"
	"checkinkeeper/pkg/idgen"
)

// CreateUser 创建用户。
func (s *Service) CreateUser(input model.User) (*model.User, error) {
	input.ID = ""
	input.Points = 0
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateUser(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建用户 %s", input.Username)
	return &input, nil
}

// GetUser 查询用户详情。
func (s *Service) GetUser(id string) (*model.User, error) {
	return s.store.GetUser(id)
}

// ListUsers 分页查询用户列表，按积分降序。
func (s *Service) ListUsers(filter model.UserFilter, page, size int) ([]*model.User, int, error) {
	all := s.store.ListUsers()
	matched := make([]*model.User, 0, len(all))
	for _, u := range all {
		if filter.Match(u) {
			matched = append(matched, u)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Points != matched[j].Points {
			return matched[i].Points > matched[j].Points
		}
		return matched[i].Username < matched[j].Username
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.User{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateUser 更新用户字段。
func (s *Service) UpdateUser(id, nickname, status string) (*model.User, error) {
	u, err := s.store.GetUser(id)
	if err != nil {
		return nil, err
	}
	if nickname != "" {
		u.Nickname = nickname
	}
	if status != "" {
		u.Status = status
	}
	u.UpdatedAt = time.Now()
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

// DeleteUser 删除用户；若已有签到记录则拒绝删除。
func (s *Service) DeleteUser(id string) error {
	if _, err := s.store.GetUser(id); err != nil {
		return err
	}
	for _, c := range s.store.ListCheckinRecords() {
		if c.UserID == id {
			return store.ErrConflict
		}
	}
	return s.store.DeleteUser(id)
}
