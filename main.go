package main

import (
	"log"
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/unrolled/secure"
	"gopkg.in/tylerb/graceful.v1"
)

const TokenHeader = "X-SalsaFlow-Token"

func main() {
	if err := run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func run() (err error) {
	defer recoverEnvironPanic(&err)

	var (
		addr        = ":" + mustGetenv("PORT")
		redisURL    = mustGetenv("REDISCLOUD_URL")
		accessToken = mustGetenv("ACCESS_TOKEN")
	)

	// Connect to Redis.
	conn, err := redis.Dial("tcp", redisURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Top-level router.
	router := mux.NewRouter()

	// Commits.
	commits := router.PathPrefix("/commits/{sha:[0-9a-f]+}")
	commits.Methods("GET").Handler(getMetadata(conn))
	commits.Methods("POST").Handler(postMetadata(conn))

	// Negroni middleware.
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())

	n.UseFunc(secure.New(secure.Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	}).HandlerFuncWithNext)

	n.UseFunc(tokenMiddleware(accessToken))
	n.UseHandler(router)

	// Start the server using graceful.
	graceful.Run(addr, 3*time.Second, n)
	return nil
}

func tokenMiddleware(accessToken string) negroni.Handler {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Make sure the token header matches.
			if token := r.Header.Get(TokenHeader); token != accessToken {
				httpError(rw, http.StatusForbidden)
				return
			}

			// Proceed.
			next(rw, r)
		})
}

func httpError(rw http.ResponseWriter, code int) {
	http.Error(rw, http.StatusText(code), code)
}
