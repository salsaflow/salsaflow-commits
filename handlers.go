package main

import (
	"bytes"
	"io"
	"net/http"
	"regexp"

	"github.com/garyburd/redigo/redis"
)

func getMetadata(conn redis.Conn) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sha := r.URL.Query().Get(":sha")
		if !isSHA(sha) {
			httpStatus(rw, http.StatusNotFound)
			return
		}

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
		sha := r.URL.Query().Get(":sha")
		if !isSHA(sha) {
			httpStatus(rw, http.StatusNotFound)
			return
		}

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

func isSHA(sha string) bool {
	return regexp.MustCompile("^[0-9a-f]{40}$").MatchString(sha)
}
