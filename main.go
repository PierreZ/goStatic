// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

var (
	// Def of flags
	portPtr = flag.Int("p", 8043, "The listening port")
	path    = flag.String("static", "/srv/http", "The path for the static files")
)

func main() {

	flag.Parse()

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	fs := http.FileServer(http.Dir(*path))
	http.Handle("/", fs)

	log.Println("Listening...")
	log.Fatalln(http.ListenAndServe(port, nil))
}
