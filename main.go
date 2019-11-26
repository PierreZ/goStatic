// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	// Def of flags
	portPtr                  = flag.Int("port", 8043, "The listening port")
	context                  = flag.String("context", "", "The 'context' path on which files are served, e.g. 'doc' will serve the files at 'http://localhost:<port>/doc/'")
	basePath                 = flag.String("path", "/srv/http", "The path for the static files")
	fallbackPath             = flag.String("fallback", "", "Default fallback file. Either absolute for a specific asset (/index.html), or relative to recursively resolve (index.html)")
	headerFlag               = flag.String("append-header", "", "HTTP response header, specified as `HeaderName:Value` that should be added to all responses.")
	basicAuth                = flag.Bool("enable-basic-auth", false, "Enable basic auth. By default, password are randomly generated. Use --set-basic-auth to set it.")
	healthCheck              = flag.Bool("enable-health", false, "Enable health check endpoint. You can call /health to get a 200 response. Useful for Kubernetes, OpenFaas, etc.")
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

	var fileSystem http.FileSystem = http.Dir(*basePath)

	if *fallbackPath != "" {
		fileSystem = fallback{
			defaultPath: *fallbackPath,
			fs:          fileSystem,
		}
	}

	handler := http.FileServer(fileSystem)

	pathPrefix := "/"
	if len(*context) > 0 {
		pathPrefix = "/" + *context + "/"
		handler = http.StripPrefix(pathPrefix, handler)
	}

	if *basicAuth {
		log.Println("Enabling Basic Auth")
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

	if *healthCheck {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Ok")
		})
	}

	http.Handle(pathPrefix, handler)

	log.Printf("Listening at 0.0.0.0%v %v...", port, pathPrefix)
	log.Fatalln(http.ListenAndServe(port, nil))
}
