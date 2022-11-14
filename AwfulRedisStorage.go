package main

import (
	"sync"
	"time"
)

type dataRecord struct {
	value string
	ttl   int
}

type AwfulRedisStorage struct {
	data  map[string]dataRecord
	mutex *sync.RWMutex
}

func (storage *AwfulRedisStorage) Get(key string) (string, bool) {
	storage.mutex.RLock()
	rec, ok := storage.data[key]
	storage.mutex.RUnlock()
	toDelete := false
	if ok && rec.ttl < int(time.Now().Unix()) {
		ok = false
	}
	value := ""
	if ok {
		value = rec.value
	}
	if toDelete {
		storage.mutex.Lock()
		delete(storage.data, key)
		storage.mutex.Unlock()
	}
	return value, ok
}

func (storage *AwfulRedisStorage) Set(key, value string, ttl int) (string, bool) {
	storage.mutex.Lock()
	rec, ok := storage.data[key]
	storage.data[key] = dataRecord{
		value: value,
		ttl:   ttl,
	}
	storage.mutex.Unlock()

	if ok && rec.ttl < int(time.Now().Unix()) {
		ok = false
	}
	old_value := ""
	if ok {
		old_value = rec.value
	}
	return old_value, ok
}

func (storage *AwfulRedisStorage) Delete(key string) (string, bool) {
	storage.mutex.Lock()
	rec, ok := storage.data[key]
	delete(storage.data, key)
	storage.mutex.Unlock()
	if ok && rec.ttl < int(time.Now().Unix()) {
		ok = false
	}
	old_value := ""
	if ok {
		old_value = rec.value
	}
	return old_value, ok
}

func (storage *AwfulRedisStorage) Keys(pattern string) []string {
	keys := []string{}
	forDelete := []string{}
	timestamp := int(time.Now().Unix())
	storage.mutex.RLock()
	for k, rec := range storage.data {
		if rec.ttl < timestamp {
			forDelete = append(forDelete, k)
			continue
		}
		if len(pattern) == 0 {
			keys = append(keys, k)
		} else if isMatch, _ := Match(pattern, k); isMatch {
			keys = append(keys, k)
		}
	}
	storage.mutex.RUnlock()

	if len(forDelete) > 0 {
		storage.mutex.Lock()
		for _, k := range forDelete {
			delete(storage.data, k)
		}
		storage.mutex.Unlock()
	}

	return keys
}

func NewAwfulRedisStorage() *AwfulRedisStorage {
	storage := new(AwfulRedisStorage)
	storage.data = map[string]dataRecord{}
	storage.mutex = &sync.RWMutex{}

	return storage
}
