// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	logLevel                 = flag.String("log-level", "info", "default: info - What level of logging to run, debug logs all requests (error, warn, info, debug)")
	logRequest               = flag.Bool("enable-logging", false, "Enable log request. NOTE: Deprecated, set log-level to debug to log all requests")
	httpsPromote             = flag.Bool("https-promote", false, "All HTTP requests should be redirected to HTTPS")
	headerConfigPath         = flag.String("header-config-path", "/config/headerConfig.json", "Path to the config file for custom response headers")
	basicAuthUser            = flag.String("basic-auth-user", "", "Username for basic auth")
	basicAuthPass            = flag.String("basic-auth-pass", "", "Password for basic auth")

	username string
	password string
)

func setupLogger(logLevel string) {
	switch logLevel {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

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

var gzPool = sync.Pool{
	New: func() interface{} {
		w := gzip.NewWriter(ioutil.Discard)
		return w
	},
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func handleReq(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *httpsPromote && r.Header.Get("X-Forwarded-Proto") == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			log.Debug().Int("http_code", 301).Str("Method", r.Method).Str("Path", r.URL.Path).Msg("Request Redirected")
			return
		}
		log.Debug().Str("Method", r.Method).Str("Path", r.URL.Path).Msg("Request Handled")
		h.ServeHTTP(w, r)
	})
}

func main() {

	flag.Parse()

	// Setting up logging output before setting up level to print out deprecation warnings
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set a pretty console output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if *logRequest {
		log.Warn().Msg("enable-logging is deprecated in favor of log-level")
	}

	//if log-level not default and enable-logging set to true, use log-level
	if *logLevel != "info" && *logRequest {
		log.Warn().Msg("log-level not 'info' and enable-logging set, using log-level")
		*logRequest = false
	}

	//if log-level is info and we have enable-logging, set log-level to debug
	if *logLevel == "info" && *logRequest {
		log.Warn().Msg("since enable-logging is set, setting log-level to debug")
		*logLevel = "debug"
	}

	//setting up the logger
	setupLogger(*logLevel)
	log.Debug().Str("Logging Level", zerolog.GlobalLevel().String()).Msg("Logger setup...")

	// sanity check
	if len(*setBasicAuth) != 0 && !*basicAuth {
		log.Debug().Msg("Basic Auth Set")
		*basicAuth = true
	}

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)

	var fileSystem http.FileSystem = http.Dir(*basePath)
	log.Debug().Str("path", *basePath).Msg("File serve path set")

	if *fallbackPath != "" {
		log.Debug().Str("FallbackPath", *fallbackPath).Msg("Fallback path set")
		fileSystem = fallback{
			defaultPath: *fallbackPath,
			fs:          fileSystem,
		}
	}

	handler := handleReq(http.FileServer(fileSystem))

	pathPrefix := "/"
	if len(*context) > 0 {
		pathPrefix = "/" + *context + "/"
		handler = http.StripPrefix(pathPrefix, handler)
	}

	if *basicAuth {
		log.Debug().Msg("Enabling Basic Auth")
		if len(*basicAuthUser) != 0 && len(*basicAuthPass) != 0 {
			parseAuth(*basicAuthUser + ":" + *basicAuthPass)
		} else if len(*setBasicAuth) != 0 {
			parseAuth(*setBasicAuth)
		} else {
			generateRandomAuth()
		}
		handler = authMiddleware(handler)
	}

	headerConfigValid := initHeaderConfig(*headerConfigPath)
	if headerConfigValid {
		handler = customHeadersMiddleware(handler)
	}

	// Extra headers.
	if len(*headerFlag) > 0 {
		header, headerValue := parseHeaderFlag(*headerFlag)
		if len(header) > 0 && len(headerValue) > 0 {
			fileServer := handler
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Debug().Str("URL", r.URL.Path).Str("header", header).Str("headerValue", headerValue).Msg("Extra Headers Handled")
				w.Header().Set(header, headerValue)
				if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
					fileServer.ServeHTTP(w, r)
				} else {
					w.Header().Set("Content-Encoding", "gzip")
					gz := gzPool.Get().(*gzip.Writer)
					defer gzPool.Put(gz)

					gz.Reset(w)
					defer gz.Close()
					fileServer.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)

				}

			})
		} else {
			log.Warn().Msg("appendHeader misconfigured; ignoring.")
		}
	}

	if *healthCheck {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			log.Debug().Msg("Returning Service Health")
			fmt.Fprintf(w, "Ok")
		})
	}

	http.Handle(pathPrefix, handler)

	log.Info().Msgf("Listening at http://0.0.0.0%v %v...", port, pathPrefix)
	if err := http.ListenAndServe(port, nil); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server startup failed")
	}

}
