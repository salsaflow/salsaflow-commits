package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bmizerany/pat"
	"github.com/codegangsta/negroni"
	"github.com/garyburd/redigo/redis"
	"github.com/unrolled/secure"
	"gopkg.in/tylerb/graceful.v1"
)

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
		isDevelopment = os.Getenv("DEVEL") != ""
		rediscloudURL = mustGetenv("REDISCLOUD_URL")
		accessToken   = mustGetenv("ACCESS_TOKEN")
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

	redisConn, err := redis.Dial("tcp", redisURL.Host)
	if err != nil {
		return err
	}
	defer redisConn.Close()

	if redisURL.User != nil {
		redisPwd, _ := redisURL.User.Password()
		if _, err := redisConn.Do("AUTH", redisPwd); err != nil {
			return err
		}
	}

	// Mux.
	mux := pat.New()

	// Commits.
	mux.Get("/commits", getMetadataBatch(redisConn))
	mux.Get("/commits/:sha", getMetadata(redisConn))

	mux.Post("/commits", postMetadataBatch(redisConn))
	mux.Post("/commits/:sha", postMetadata(redisConn))

	// Negroni middleware.
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())

	n.UseFunc(secure.New(secure.Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		IsDevelopment:   isDevelopment,
	}).HandlerFuncWithNext)

	n.Use(bearerTokenMiddleware(accessToken))
	n.UseHandler(mux)

	// Start the server using graceful.
	graceful.Run(addr, 3*time.Second, n)
	return nil
}

func bearerTokenMiddleware(accessToken string) negroni.Handler {
	authorizationHeaderValue := "Bearer " + accessToken

	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Make sure the token header matches.
			if header := r.Header.Get("Authorization"); header != authorizationHeaderValue {
				httpStatus(rw, http.StatusForbidden)
				return
			}

			// Proceed.
			next(rw, r)
		})
}
