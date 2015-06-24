package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
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
		host          = os.Getenv("HOST")
		port          = os.Getenv("PORT")
		rediscloudURL = mustGetenv("REDISCLOUD_URL")
		accessToken   = mustGetenv("ACCESS_TOKEN")
		isDevelopment = os.Getenv("DEVEL") != ""
	)

	// Get the listening address.
	// When testing on localhost, it's enough to set HOST.
	// On the other hand, Heroku is setting PORT.
	var addr string
	switch {
	case host != "":
		addr = host
	case port != "":
		addr = ":" + port
	default:
		panic(&ErrVarNotSet{"PORT"})
	}

	// Connect to Redis.
	redisURL, err := url.Parse(rediscloudURL)
	if err != nil {
		return err
	}

	conn, err := redis.Dial("tcp", redisURL.Host)
	if err != nil {
		return err
	}
	defer conn.Close()

	if redisURL.User != nil {
		redisPwd, _ := redisURL.User.Password()
		if _, err := conn.Do("AUTH", redisPwd); err != nil {
			return err
		}
	}

	// Router.
	router := mux.NewRouter()

	// Commits.
	router.Handle("/commits", getMetadataBatch(conn))

	/*
		commits := router.PathPrefix("/commits")
		commits.Methods("GET").Handler(getMetadataBatch(conn))
		commits.Methods("POST").Handler(postMetadataBatch(conn))

		commit := commits.PathPrefix("/{sha:[0-9a-f]{40}}")
		commit.Methods("GET").Handler(getMetadata(conn))
		commit.Methods("POST").Handler(postMetadata(conn))
	*/

	// Negroni middleware.
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())

	n.UseFunc(secure.New(secure.Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		IsDevelopment:   isDevelopment,
	}).HandlerFuncWithNext)

	n.Use(tokenMiddleware(accessToken))
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
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			// Proceed.
			next(rw, r)
		})
}
