/*
 * My awful Redis
 *
 * API version: 0.1.0
 */
package main

import (
	"log"
	"net/http"
)

func main() {
	storage := NewAwfulRedisStorage()

	handler := NewAwfulRedisHandler(storage)

	log.Println("starting server at :8082")
	http.ListenAndServe(":8082", handler)
}
