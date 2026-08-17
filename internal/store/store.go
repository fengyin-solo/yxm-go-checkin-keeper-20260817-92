// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"checkinkeeper/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// 用户
	CreateUser(u *model.User) error
	GetUser(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	ListUsers() []*model.User
	UpdateUser(u *model.User) error
	DeleteUser(id string) error

	// 签到活动
	CreateActivity(a *model.Activity) error
	GetActivity(id string) (*model.Activity, error)
	GetActivityByCode(code string) (*model.Activity, error)
	ListActivities() []*model.Activity
	UpdateActivity(a *model.Activity) error
	DeleteActivity(id string) error

	// 签到记录
	CreateCheckinRecord(c *model.CheckinRecord) error
	GetCheckinRecord(id string) (*model.CheckinRecord, error)
	ListCheckinRecords() []*model.CheckinRecord
	GetCheckinByUserActivityDate(userID, activityID, date string) (*model.CheckinRecord, error)
	LatestCheckinByUserActivity(userID, activityID string) (*model.CheckinRecord, error)

	// 奖励规则
	CreateRewardRule(r *model.RewardRule) error
	GetRewardRule(id string) (*model.RewardRule, error)
	ListRewardRules() []*model.RewardRule
	UpdateRewardRule(r *model.RewardRule) error
	DeleteRewardRule(id string) error

	// 奖励发放
	CreateRewardGrant(g *model.RewardGrant) error
	GetRewardGrant(id string) (*model.RewardGrant, error)
	ListRewardGrants() []*model.RewardGrant
}
