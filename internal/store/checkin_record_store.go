package store

import (
	"checkinkeeper/internal/model"
)

func (s *MemoryStore) CreateCheckinRecord(c *model.CheckinRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.checkins {
		if exist.UserID == c.UserID && exist.ActivityID == c.ActivityID &&
			exist.CheckinDate == c.CheckinDate {
			return ErrConflict
		}
	}
	s.checkins[c.ID] = c
	return nil
}

func (s *MemoryStore) GetCheckinRecord(id string) (*model.CheckinRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.checkins[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *MemoryStore) ListCheckinRecords() []*model.CheckinRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.CheckinRecord, 0, len(s.checkins))
	for _, c := range s.checkins {
		list = append(list, c)
	}
	return list
}

// GetCheckinByUserActivityDate 查询某用户在某活动某天的签到记录。
func (s *MemoryStore) GetCheckinByUserActivityDate(userID, activityID, date string) (*model.CheckinRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.checkins {
		if c.UserID == userID && c.ActivityID == activityID && c.CheckinDate == date {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

// LatestCheckinByUserActivity 查询某用户在某活动最近一次签到记录（按日期最大）。
func (s *MemoryStore) LatestCheckinByUserActivity(userID, activityID string) (*model.CheckinRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var latest *model.CheckinRecord
	for _, c := range s.checkins {
		if c.UserID != userID || c.ActivityID != activityID {
			continue
		}
		if latest == nil || c.CheckinDate > latest.CheckinDate {
			latest = c
		}
	}
	if latest == nil {
		return nil, ErrNotFound
	}
	return latest, nil
}
