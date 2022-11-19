/*
Пакет с имплементацией key-value хранилища AwfulRedisStorage
Под капотом это обычный hashmap
TTL проверяется в момент отдачи значения или в момент построения списка ключей
Структура потокобезопасная (содержит Mutex)

Объект нужно создавать через вызов конструктора
obj := NewAwfulRedisStorage
*/
package main

import (
	"sync"
	"time"
)

type dataRecord struct {
	value string
	ttl   int
}

// Интерфейс к хранилищу
type AwfulRedisStorage struct {
	data  map[string]dataRecord
	mutex *sync.RWMutex
}

// Метод возвращает значение по ключу
// Если ключ не найден, то возвращается пустая строка и false
func (storage *AwfulRedisStorage) Get(key string) (string, bool) {
	storage.mutex.RLock()
	rec, ok := storage.data[key]
	storage.mutex.RUnlock()
	toDelete := false
	if ok && rec.ttl < int(time.Now().Unix()) {
		ok = false
		toDelete = true
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

// Метод сохраняет значение value по ключу key, которое удалится в ttl (unix timestamp в секундах)
// Если такой ключ уже был и не устарел, то возвращается его значение и true
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

// Метод удаляет значение value по ключу key
// Если такой ключ уже был и не устарел, то возвращается его значение и true
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

// Метод возвращает список актуальных ключей
func (storage *AwfulRedisStorage) Keys(pattern string) ([]string, error) {
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
		} else {
			isMatch, err := Match(pattern, k)
			if err != nil {
				return []string{}, err
			}
			if isMatch {
				keys = append(keys, k)
			}
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

	// sort.Strings(keys)
	return keys, nil
}

// Конструктор, возвращает ссылку на инициализированный объект AwfulRedisStorage
func NewAwfulRedisStorage() *AwfulRedisStorage {
	storage := new(AwfulRedisStorage)
	storage.data = map[string]dataRecord{}
	storage.mutex = &sync.RWMutex{}

	return storage
}
