/*
 * My awful Redis
 *
 * API version: 0.1.0
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	ip := flag.String("ip", "", "ip to accept connections, default any")
	port := flag.Int("p", 8080, "port to accept connections")
	flag.Parse()
	ipPort := fmt.Sprintf("%s:%d", *ip, *port)
	storage := NewAwfulRedisStorage()
	handler := NewAwfulRedisHandler(storage)

	log.Println("starting server at", ipPort)
	http.ListenAndServe(ipPort, handler)
}
