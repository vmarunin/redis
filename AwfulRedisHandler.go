/*
Пакет с имплементацией http.Handler RESTful сервиса AwfulRedis
Описание API лежит в openapi.yaml

Объект нужно создавать через вызов конструктора
obj := NewAwfulRedisHandler
*/
package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

// Интерфейс к handler, ничего не экспортируем
type AwfulRedisHandler struct {
	storage *AwfulRedisStorage
	prefix  string
}

// Конструктор
// Принимает на вход ссылку на объект хранилища
func NewAwfulRedisHandler(storage *AwfulRedisStorage) http.Handler {
	handler := new(AwfulRedisHandler)
	handler.storage = storage
	handler.prefix = "/redis/v1"

	mux := http.NewServeMux()
	mux.HandleFunc(handler.prefix+"/keys", handler.ProcessKeys)
	mux.HandleFunc(handler.prefix+"/key/", handler.ProcessKey)
	return mux
}

// Обработчик пути вида /keys
func (handler *AwfulRedisHandler) ProcessKeys(w http.ResponseWriter, r *http.Request) {
	log.Println("ProcessKeys", r.Method, r.URL.Path)
	// if r.URL.Path != handler.prefix+"/keys" {
	// 	handler.responseError(http.StatusBadRequest, "Bad path for ProcessKeys", w)
	// 	return
	// }
	if r.Method == "OPTIONS" {
		handler.responseOK(nil, http.StatusOK, w)
		return
	}
	if r.Method != "GET" {
		handler.responseError(http.StatusBadRequest, "Method not allowed", w)
		return
	}
	pattern := r.FormValue("pattern")
	keys, err := handler.storage.Keys(pattern)
	if err != nil {
		handler.responseError(http.StatusBadRequest, err.Error(), w)
		return
	}

	handler.responseOK(keys, http.StatusOK, w)
}

// Обработчик пути вида /key/{id}
func (handler *AwfulRedisHandler) ProcessKey(w http.ResponseWriter, r *http.Request) {
	log.Println("ProcessKey", r.Method, r.URL.Path)
	// if !strings.HasPrefix(r.URL.Path, handler.prefix+"/key/") {
	// 	handler.responseError(http.StatusBadRequest, "Bad path for ProcessKey", w)
	// 	return
	// }
	if r.Method == "OPTIONS" {
		handler.responseOK(nil, http.StatusOK, w)
		return
	}
	prefixLen := len(handler.prefix) + 5 // handler.prefix+"/key/"
	key := r.URL.Path[prefixLen:]
	if r.Method == "GET" {
		value, ok := handler.storage.Get(key)
		respData := map[string]interface{}{}
		respData["value"] = value
		respData["ok"] = ok
		handler.responseOK(respData, http.StatusOK, w)
		return
	}
	if r.Method == "DELETE" {
		log.Println("Before")
		value, ok := handler.storage.Delete(key)
		log.Println("After")
		respData := map[string]interface{}{}
		respData["value"] = value
		respData["ok"] = ok
		handler.responseOK(respData, http.StatusOK, w)
		return
	}
	if r.Method == "PUT" {
		reqData := map[string]interface{}{}
		jsonDecoder := json.NewDecoder(r.Body)
		jsonDecoder.UseNumber()
		err := jsonDecoder.Decode(&reqData)
		if err != nil {
			handler.responseError(http.StatusBadRequest, `can't parse input`+err.Error(), w)
			return
		}
		value := reqData["value"].(string)
		ttlI, ok := reqData["ttl"]
		ttl := math.MaxInt
		if ok {
			ttl64, err := ttlI.(json.Number).Int64()
			if err != nil {
				handler.responseError(http.StatusBadRequest, `can't parse ttl`+err.Error(), w)
				return
			}
			ttl = int(ttl64) + int(time.Now().Unix())
		}
		old_value, ok := handler.storage.Set(key, value, ttl)
		respData := map[string]interface{}{}
		respData["value"] = old_value
		respData["ok"] = ok
		handler.responseOK(respData, http.StatusOK, w)
		return
	}
	handler.responseError(http.StatusBadRequest, "Method not allowed", w)
}

func (handler *AwfulRedisHandler) responseOK(data interface{}, http_code int, w http.ResponseWriter) {
	log.Println("responseOK", http_code, data)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,DELETE")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)

	responseBytes, _ := json.Marshal(data)
	w.Write(responseBytes)
}

func (handler *AwfulRedisHandler) responseError(http_code int, err string, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,DELETE")
	w.WriteHeader(http_code)

	io.WriteString(w, err)
}
