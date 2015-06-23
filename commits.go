package main

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func getMetadata(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

	})
}

func postMetadata(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

	})
}

func getMetadataBatch(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

	})
}

func postMetadataBatch(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

	})
}
