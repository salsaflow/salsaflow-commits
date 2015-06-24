package main

import (
	"log"
	"net/http"
)

func httpStatus(rw http.ResponseWriter, code int) {
	http.Error(rw, http.StatusText(code), code)
}

func httpError(rw http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR [%v %v] - %v\n", r.Method, r.URL.Path, err)
	httpStatus(rw, http.StatusInternalServerError)
}

func httpBadRequest(rw http.ResponseWriter, status string) {
	http.Error(rw, status, http.StatusBadRequest)
}
