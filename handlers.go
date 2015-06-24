package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getMetadata(c *mgo.Collection) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Get the commit SHA.
		sha := r.URL.Query().Get(":sha")
		if !isSHA(sha) {
			httpStatus(rw, http.StatusNotFound)
			return
		}

		// Fetch the associated DB record.
		var commit map[string]interface{}
		err := c.Find(bson.M{
			CommitShaField: sha,
		}).One(&commit)
		if err != nil {
			if err == mgo.ErrNotFound {
				httpStatus(rw, http.StatusNotFound)
				return
			}
			httpError(rw, r, err)
			return
		}

		// Return the record.
		delete(commit, "_id")
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(commit)
	})
}

func postMetadata(c *mgo.Collection) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Parse the request.
		var commits []map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&commits); err != nil {
			httpBadRequest(rw, "Not an array of commit objects")
			return
		}

		// Make sure the commit SHAs are there.
		for _, commit := range commits {
			v, ok := commit[CommitShaField]
			if !ok {
				httpBadRequest(rw, fmt.Sprintf("key '%v' is missing", CommitShaField))
				return
			}
			sha, ok := v.(string)
			if !ok {
				httpBadRequest(rw, fmt.Sprintf("key '%v': not a string", CommitShaField))
				return
			}
			if !isSHA(sha) {
				httpBadRequest(rw, fmt.Sprintf("key '%v': not a valid commit SHA", CommitShaField))
				return
			}
		}

		// Store the commit records.
		for _, commit := range commits {
			sha := commit[CommitShaField].(string)

			// Write the commit record into the database.
			if _, err := c.Upsert(bson.M{CommitShaField: sha}, commit); err != nil {
				httpError(rw, r, err)
				return
			}
		}

		// Return 202 Accepted.
		httpStatus(rw, http.StatusAccepted)
	})
}

func isSHA(sha string) bool {
	return regexp.MustCompile("^[0-9a-f]{40}$").MatchString(sha)
}
