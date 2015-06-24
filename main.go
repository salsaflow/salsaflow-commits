package main

import (
	"log"
	"net/http"
	"time"

	"github.com/bmizerany/pat"
	"github.com/codegangsta/negroni"
	"github.com/unrolled/secure"
	"gopkg.in/mgo.v2"
	"gopkg.in/tylerb/graceful.v1"
)

const (
	MongoDatabaseName   = ""
	MongoCollectionName = "commits"
	CommitShaField      = "commit_sha"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func run() (err error) {
	// Load server config.
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// Connect to MongoDB.
	mgoSession, err := mgo.Dial(config.MongoURL)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	mgoSession.SetSafe(&mgo.Safe{
		WMode: "majority",
	})

	collection := mgoSession.DB(MongoDatabaseName).C(MongoCollectionName)

	// Ensure index.
	err = collection.EnsureIndex(mgo.Index{
		Name:       "commit hashes",
		Key:        []string{CommitShaField},
		Unique:     true,
		Background: true,
	})
	if err != nil {
		return err
	}

	// Mux.
	mux := pat.New()

	// Commits.
	mux.Get("/commits/:sha", getMetadata(collection))
	mux.Post("/commits", postMetadata(collection))

	// Negroni middleware.
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())

	n.UseFunc(secure.New(secure.Options{
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		IsDevelopment:   config.IsDevelopment,
	}).HandlerFuncWithNext)

	n.Use(bearerTokenMiddleware(config.AccessToken))
	n.UseHandler(mux)

	// Start the server using graceful.
	graceful.Run(config.Addr, 3*time.Second, n)
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
