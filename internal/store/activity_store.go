package store

import (
	"checkinkeeper/internal/model"
)

func (s *MemoryStore) CreateActivity(a *model.Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.activities {
		if exist.Code == a.Code {
			return ErrConflict
		}
	}
	s.activities[a.ID] = a
	return nil
}

func (s *MemoryStore) GetActivity(id string) (*model.Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.activities[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) GetActivityByCode(code string) (*model.Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.activities {
		if a.Code == code {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListActivities() []*model.Activity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Activity, 0, len(s.activities))
	for _, a := range s.activities {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateActivity(a *model.Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.activities[a.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.activities {
		if exist.ID != a.ID && exist.Code == a.Code {
			return ErrConflict
		}
	}
	s.activities[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteActivity(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.activities[id]; !ok {
		return ErrNotFound
	}
	delete(s.activities, id)
	return nil
}
