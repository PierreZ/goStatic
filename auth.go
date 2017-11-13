package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// authMiddleware checks basic auth
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if strings.Compare(pair[0], username) != 0 || strings.Compare(pair[1], password) != 0 {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseAuth(auth string) {
	identity := strings.Split(*setBasicAuth, ":")
	if len(identity) != 2 {
		log.Fatalln("basic auth must be like this: user:password")
	}

	username = identity[0]
	password = identity[1]
}

func generateRandomAuth() {
	username = "gopher"
	password = generateRandomString()
	log.Printf("User generated for basic auth. User:'%v', password:'%v'\n", username, password)
}

func generateRandomString() string {

	b := make([]byte, *sizeRandom)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}
