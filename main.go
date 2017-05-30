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
	portPtr    = flag.Int("p", 8043, "The listening port")
	path       = flag.String("static", "/srv/http", "The path for the static files")
	headerFlag = flag.String("appendHeader", "", "HTTP response header, specified as `HeaderName:Value` that should be added to all responses.")
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

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	handler := http.FileServer(http.Dir(*path))

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
