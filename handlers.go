package main

import (
	"bytes"
	"encoding/json"
	"io"
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
		var payload bytes.Buffer
		if err := json.NewEncoder(&payload).Encode(commit); err != nil {
			httpError(rw, r, err)
			return
		}
		io.Copy(rw, &payload)
	})
}

func postMetadata(c *mgo.Collection) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Parse the request.
		var commits []map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&commits); err != nil {
			httpStatus(rw, http.StatusBadRequest)
			return
		}

		// Make sure the commit SHAs are there.
		for _, commit := range commits {
			v, ok := commit[CommitShaField]
			if !ok {
				httpStatus(rw, http.StatusBadRequest)
				return
			}
			sha, ok := v.(string)
			if !ok {
				httpStatus(rw, http.StatusBadRequest)
				return
			}
			if !isSHA(sha) {
				httpStatus(rw, http.StatusBadRequest)
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
