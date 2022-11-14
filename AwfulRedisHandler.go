package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AwfulRedisHandler struct {
	storage *AwfulRedisStorage
	prefix  string
}

func NewAwfulRedisHandler(storage *AwfulRedisStorage) http.Handler {
	handler := new(AwfulRedisHandler)
	handler.storage = storage
	handler.prefix = "/redis/v1"

	mux := http.NewServeMux()
	mux.HandleFunc(handler.prefix+"/keys", handler.ProcessKeys)
	mux.HandleFunc(handler.prefix+"/key/", handler.ProcessKey)
	return mux
}

func (handler *AwfulRedisHandler) ProcessKeys(w http.ResponseWriter, r *http.Request) {
	log.Println("ProcessKeys")
	if r.URL.Path != handler.prefix+"/keys" {
		w.WriteHeader(http.StatusBadRequest)
		handler.CORSHeaders(w)
		io.WriteString(w, `{"error": "Bad path for keys"}`)
		return
	}
	if r.Method == "OPTIONS" {
		handler.CORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		w.WriteHeader(http.StatusBadRequest)
		handler.CORSHeaders(w)
		io.WriteString(w, `{"error": "Only HTTP GET allowed"}`)
		return
	}
	pattern := r.FormValue("pattern")
	keys := handler.storage.Keys(pattern)
	handler.CORSHeaders(w)
	w.WriteHeader(http.StatusOK)
	responseBytes, _ := json.Marshal(keys)
	w.Write(responseBytes)
}

func (handler *AwfulRedisHandler) ProcessKey(w http.ResponseWriter, r *http.Request) {
	log.Println("ProcessKey", r.Method, r.URL.Path)
	if !strings.HasPrefix(r.URL.Path, handler.prefix+"/key/") {
		handler.CORSHeaders(w)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error": "Bad path for keys"}`)
		return
	}
	prefixLen := len(handler.prefix) + 5 // handler.prefix+"/key/"
	key := r.URL.Path[prefixLen:]
	if r.Method == "GET" {
		value, ok := handler.storage.Get(key)
		if ok {
			handler.CORSHeaders(w)
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, value)
			return
		} else {
			handler.CORSHeaders(w)
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `Key not found`)
			return
		}
	}
	if r.Method == "DELETE" {
		value, ok := handler.storage.Delete(key)
		if !ok {
			value = ""
		}
		handler.CORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, value)
		return
	}
	if r.Method == "PUT" {
		value := r.FormValue("value")
		ttlStr := r.FormValue("ttl")
		ttl := math.MaxInt
		if ttlStr != "" {
			ttlInt, err := strconv.Atoi(ttlStr)
			if err != nil {
				handler.CORSHeaders(w)
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"error": "Cant't parse ttl"}`)
			}
			ttl = ttlInt + int(time.Now().Unix())
		}
		old_value, ok := handler.storage.Set(key, value, ttl)
		if !ok {
			old_value = ""
		}
		handler.CORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, old_value)
		return
	}
	if r.Method == "OPTIONS" {
		handler.CORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"error": "HTTP Method unknown"}`)
}

func (handler *AwfulRedisHandler) CORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,DELETE")
}
