package tgw

import (
	"encoding/json"
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
)

//==================================================
type memcachedSessionStore struct {
	client *memcache.Client
}

func NewMemcachedSessionStore(mc string) *memcachedSessionStore {

	return &memcachedSessionStore{client: memcache.New(mc)}
}

func memcache_key(id, key string) string {
	return id + key
}

func (s *memcachedSessionStore) Clear(id string, key string) {

	s.client.Delete(memcache_key(id, key))
}

func (s *memcachedSessionStore) Get(id string, key string, val interface{}) (err error) {

	item, err := s.client.Get(memcache_key(id, key))
	if err != nil {
		return
	}
	if len(item.Value) == 0 {
		err = errors.New("memcache miss cache,key:" + memcache_key(id, key))
		return
	}
	err = json.Unmarshal(item.Value, &val)
	log.Println("json.Unmarshal:err", err, "val:", val)
	if err != nil {
		return
	}
	return
}

func (s *memcachedSessionStore) Set(id string, key string, val interface{}) (err error) {

	bval, err := json.Marshal(val)
	if err != nil {
		return
	}

	item := &memcache.Item{
		Key:   memcache_key(id, key),
		Value: bval,
	}
	err = s.client.Set(item)
	return
}
