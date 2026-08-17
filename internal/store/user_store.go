package store

import (
	"checkinkeeper/internal/model"
)

func (s *MemoryStore) CreateUser(u *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.users {
		if exist.Username == u.Username {
			return ErrConflict
		}
	}
	s.users[u.ID] = u
	return nil
}

func (s *MemoryStore) GetUser(id string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (s *MemoryStore) GetUserByUsername(username string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListUsers() []*model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.User, 0, len(s.users))
	for _, u := range s.users {
		list = append(list, u)
	}
	return list
}

func (s *MemoryStore) UpdateUser(u *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.users {
		if exist.ID != u.ID && exist.Username == u.Username {
			return ErrConflict
		}
	}
	s.users[u.ID] = u
	return nil
}

func (s *MemoryStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return ErrNotFound
	}
	delete(s.users, id)
	return nil
}
