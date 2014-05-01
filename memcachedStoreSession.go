package tgw

import (
	"github.com/icattlecoder/mcClient"
)

//==================================================
type memcachedSessionStore struct {
	client mcClient.MC
}

func NewMemcachedSessionStore(mc ...string) *memcachedSessionStore {
	return &memcachedSessionStore{client: mcClient.NewGobMCClient("session", mc...)}
}

func memcache_key(id, key string) string {
	return id + key
}

func (s *memcachedSessionStore) Clear(id string, key string) {

	s.client.Delete(memcache_key(id, key))
}

func (s *memcachedSessionStore) Get(id string, key string, val interface{}) (err error) {

	return s.client.Get(memcache_key(id, key), val)
}

func (s *memcachedSessionStore) GetString(id string, key string) (val string, err error) {

	return s.client.GetString(memcache_key(id, key))
}

func (s *memcachedSessionStore) Set(id string, key string, val interface{}) (err error) {

	return s.client.Set(memcache_key(id, key), val)
}

func (s *memcachedSessionStore) SetString(id string, key string, val string) (err error) {

	return s.client.SetString(memcache_key(id, key), val)
}
