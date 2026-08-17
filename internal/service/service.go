// Package service 实现业务逻辑层。
package service

import (
	"checkinkeeper/internal/config"
	"checkinkeeper/internal/store"
	"checkinkeeper/pkg/logger"
)

// Service 聚合全部业务方法，依赖 Store 做数据访问。
type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}
