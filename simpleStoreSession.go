package tgw

import (
	"errors"
)

//==================================================
type d map[string]interface{}
type D map[string]d
type simpleSessionStore struct {
	value *D
}

func NewSimpleSessionStore() *simpleSessionStore {

	return &simpleSessionStore{value: &D{}}
}

func (s *simpleSessionStore) Clear(id string, key string) {
	delete((*s.value)[id], key)
}

func (s *simpleSessionStore) Get(id string, key string, val interface{}) (err error) {

	ma, ok := (*s.value)[id]
	if !ok {
		(*s.value)[id] = make(d)
		err = errors.New("simpleSessionStore.Get error : No such SESSION_ID " + id)
		return
	}

	if _, ok := ma[key]; ok {
		val = ma[key]
	} else {
		err = errors.New("simpleSessionStore.Get error : No such Key " + key)
	}
	return
}

func (s *simpleSessionStore) Set(id string, key string, val interface{}) (err error) {
	ma, ok := (*s.value)[id]
	if !ok {
		ma = make(d)
	}
	ma[key] = val
	return
}
