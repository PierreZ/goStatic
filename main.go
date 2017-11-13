// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	// Def of flags
	portPtr                  = flag.Int("port", 8043, "The listening port")
	path                     = flag.String("path", "/srv/http", "The path for the static files")
	headerFlag               = flag.String("append-header", "", "HTTP response header, specified as `HeaderName:Value` that should be added to all responses.")
	basicAuth                = flag.Bool("enable-basic-auth", false, "Enable basic auth. By default, password are randomly generated. Use --set-basic-auth to set it.")
	setBasicAuth             = flag.String("set-basic-auth", "", "Define the basic auth. Form must be user:password")
	defaultUsernameBasicAuth = flag.String("default-user-basic-auth", "gopher", "Define the user")
	sizeRandom               = flag.Int("password-length", 16, "Size of the randomized password")

	username string
	password string
)

func parseHeaderFlag(headerFlag string) (string, string) {
	if len(headerFlag) == 0 {
		return "", ""
	}
	pieces := strings.SplitN(headerFlag, ":", 2)
	if len(pieces) == 1 {
		return pieces[0], ""
	}
	return pieces[0], pieces[1]
}

func main() {

	flag.Parse()

	// sanity check
	if len(*setBasicAuth) != 0 && !*basicAuth {
		*basicAuth = true
	}

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	handler := http.FileServer(http.Dir(*path))

	if *basicAuth {
		if len(*setBasicAuth) != 0 {
			parseAuth(*setBasicAuth)
		} else {
			generateRandomAuth()
		}
		handler = authMiddleware(handler)
	}

	// Extra headers.
	if len(*headerFlag) > 0 {
		header, headerValue := parseHeaderFlag(*headerFlag)
		if len(header) > 0 && len(headerValue) > 0 {
			fileServer := handler
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(header, headerValue)
				fileServer.ServeHTTP(w, r)
			})
		} else {
			log.Println("appendHeader misconfigured; ignoring.")
		}
	}

	http.Handle("/", handler)

	log.Printf("Listening at 0.0.0.0%v...", port)
	log.Fatalln(http.ListenAndServe(port, nil))
}
