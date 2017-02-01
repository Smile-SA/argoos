package main

import (
	"net/http"
	"testing"
)

// TestToken checks token authentication.
func TestToken(t *testing.T) {
	// no token
	r, _ := http.NewRequest("POST", "/event", nil)
	err := checkToken(r)
	if err != nil {
		t.Log("Should not return error, got", err)
	}

	// set a token, argoos should break with an error
	TOKEN = "A nice token"
	err = checkToken(r)

	if e, ok := err.(*BadTokenError); !ok {
		t.Error("Bad error type, should be BadTokenError, got", e)
	}

	// add token in header, and argoos stop to complain
	r.Header.Add("X-Argoos-Token", TOKEN)

	err = checkToken(r)
	if err != nil {
		t.Error("Should not return error, got", err)
	}
}
