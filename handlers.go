package main

import (
	"bytes"
	"io"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

func getMetadata(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sha := mux.Vars(r)["sha"]
		content, err := redis.String(conn.Do("GET", sha))
		if err != nil {
			httpError(rw, r, err)
			return
		}

		io.WriteString(rw, content)
	})
}

func postMetadata(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sha := mux.Vars(r)["sha"]

		var metadata bytes.Buffer
		if _, err := io.Copy(&metadata, r.Body); err != nil {
			httpError(rw, r, err)
			return
		}

		if _, err := conn.Do("SET", sha, metadata.String()); err != nil {
			httpError(rw, r, err)
			return
		}

		httpStatus(rw, http.StatusAccepted)
	})
}

func getMetadataBatch(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httpStatus(rw, http.StatusNotImplemented)
	})
}

func postMetadataBatch(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		httpStatus(rw, http.StatusNotImplemented)
	})
}
