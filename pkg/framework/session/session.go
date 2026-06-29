package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"lychee-go/pkg/framework/config"
	"lychee-go/pkg/framework/logger"
)

type Session struct {
	ID        string
	Data      map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore interface {
	Save(s *Session) error
	Get(id string) (*Session, error)
	Delete(id string) error
	Cleanup()
}

type MemoryStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]*Session)}
}

func (s *MemoryStore) Save(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	if time.Now().After(sess.ExpiresAt) {
		delete(s.sessions, id)
		return nil, errors.New("session expired")
	}
	return sess, nil
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

func (s *MemoryStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

type Manager struct {
	store     SessionStore
	expireSec int64
}

var (
	manager *Manager
	once    sync.Once
)

func Init() {
	once.Do(func() {
		ttl := int64(config.GetInt("session.ttl", 7200))
		manager = &Manager{
			store:     NewMemoryStore(),
			expireSec: ttl,
		}
		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				manager.store.Cleanup()
			}
		}()
		logger.Info("[Session] Initialized (ttl: %ds)", ttl)
	})
}

func SetStore(store SessionStore) {
	if manager != nil {
		manager.store = store
	}
}

func generateID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Create() (*Session, error) {
	if manager == nil {
		Init()
	}
	sess := &Session{
		ID:        generateID(),
		Data:      make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(manager.expireSec) * time.Second),
	}
	if err := manager.store.Save(sess); err != nil {
		return nil, err
	}
	logger.Debug("[Session] Created: %s", sess.ID)
	return sess, nil
}

func Get(id string) (*Session, error) {
	if manager == nil {
		Init()
	}
	return manager.store.Get(id)
}

func Delete(id string) error {
	if manager == nil {
		Init()
	}
	logger.Debug("[Session] Deleted: %s", id)
	return manager.store.Delete(id)
}

func (s *Session) Set(key string, value interface{}) {
	s.Data[key] = value
	s.UpdatedAt = time.Now()
	s.ExpiresAt = time.Now().Add(time.Duration(manager.expireSec) * time.Second)
	manager.store.Save(s)
}

func (s *Session) Get(key string) (interface{}, bool) {
	v, ok := s.Data[key]
	return v, ok
}

func (s *Session) GetString(key string) string {
	v, ok := s.Data[key]
	if !ok {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

func (s *Session) GetInt(key string) int {
	v, ok := s.Data[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	}
	return 0
}

func (s *Session) GetUint(key string) uint {
	return uint(s.GetInt(key))
}

func (s *Session) GetBool(key string) bool {
	v, ok := s.Data[key]
	if !ok {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func (s *Session) Delete(key string) {
	delete(s.Data, key)
	s.UpdatedAt = time.Now()
	manager.store.Save(s)
}

func (s *Session) Clear() {
	s.Data = make(map[string]interface{})
	s.UpdatedAt = time.Now()
	manager.store.Save(s)
}

func (s *Session) Has(key string) bool {
	_, ok := s.Data[key]
	return ok
}

func (s *Session) All() map[string]interface{} {
	return s.Data
}

func (s *Session) Flash(key string, value interface{}) {
	s.Set("__flash__"+key, value)
}

func (s *Session) GetFlash(key string) interface{} {
	flashKey := "__flash__" + key
	v, ok := s.Data[flashKey]
	if ok {
		delete(s.Data, flashKey)
		s.UpdatedAt = time.Now()
		manager.store.Save(s)
	}
	return v
}

func (s *Session) Renew() {
	s.ExpiresAt = time.Now().Add(time.Duration(manager.expireSec) * time.Second)
	s.UpdatedAt = time.Now()
	manager.store.Save(s)
}

func (s *Session) TTL() int64 {
	return int64(time.Until(s.ExpiresAt).Seconds())
}
