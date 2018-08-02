package httpapi

import (
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/oauth2"
)

// memStore provides an in-memory tokenStore implementation which is used by service_test for unit testing.
type memStore struct {
	store map[string][]byte
	lock  sync.Mutex
}

func NewMemStore() *memStore {
	return &memStore{
		store: make(map[string][]byte),
	}
}

func (s *memStore) GetToken(user User) (*oauth2.Token, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	val, ok := s.store[user.UserID()]
	if !ok {
		return nil, nil
	}
	tok := &oauth2.Token{}
	if err := json.Unmarshal(val, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func (s *memStore) SaveToken(user User, token *oauth2.Token) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if token == nil {
		return fmt.Errorf("token cannot be empty")
	}
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	s.store[user.UserID()] = data
	return nil
}
