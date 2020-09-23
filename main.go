// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"encoding/json"
	"os"
	"path/filepath"
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
	logRequest               = flag.Bool("enable-logging", false, "Enable log request")
	httpsPromote             = flag.Bool("https-promote", false, "All HTTP requests should be redirected to HTTPS")

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

func handleReq(h http.Handler, headerConfigs HeaderConfigs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
		log.Println("======================")
		log.Println("path =>", r.URL.Path)
		log.Println("extension =>", filepath.Ext(r.URL.Path))
		log.Println("pagedata =>", strings.HasPrefix(r.URL.Path, "/page-data"))
		log.Println("======================")
		*/
		if *httpsPromote && r.Header.Get("X-Forwarded-Proto") == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			if *logRequest {
				log.Println(301, r.Method, r.URL.Path)
			}
			return
		}

		if *logRequest {
			log.Println(r.Method, r.URL.Path)
		}

		// apply custom header config
		if (len(headerConfigs.Configs) > 0) {
			reqFileExtension := filepath.Ext(r.URL.Path)

			for i := 0; i < len(headerConfigs.Configs); i++ {
				configEntry := headerConfigs.Configs[i]

				fileMatch := configEntry.FileExtension == "*" || reqFileExtension == "."+ configEntry.FileExtension
				pathMatch := configEntry.Path == "*" || strings.HasPrefix(r.URL.Path, configEntry.Path)

				// log.Println("fileMatch", reqFileExtension, configEntry.FileExtension)
				// log.Println("pathMatch", r.URL.Path, configEntry.Path)

				if fileMatch && pathMatch {
					log.Println(r.URL.Path, configEntry.FileExtension, configEntry.Path)
					for j := 0; j < len(configEntry.Headers); j++ {
						headerEntry := configEntry.Headers[j]

						log.Println("add header", headerEntry.Key, headerEntry.Value, "to", r.URL.Path)
						w.Header().Set(headerEntry.Key, headerEntry.Value)
					}
				}
			}
		}

		h.ServeHTTP(w, r)
	})
}

type HeaderConfigs struct {
	Configs []HConfig `json:"configs"`
}

type HConfig struct {
	Path string `json:"path"`
	FileExtension string `json:"fileExtension"`
	Headers []Header `json:"headers"`
}

type Header struct {
	Key string `json:"key"`
	Value string `json:"value"`
}


func initHeaderConfig() HeaderConfigs {
	if fileExists("/config/headerConfig.json") {
		jsonFile, err := os.Open("/config/headerConfig.json")
		if err != nil {
			fmt.Println(err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var headerConfigs HeaderConfigs

		json.Unmarshal(byteValue, &headerConfigs)

		fmt.Println("Found header config file. Rules:")

		for i := 0; i < len(headerConfigs.Configs); i++ {
			fmt.Println("==============================================================")
			fmt.Println("Path: " + headerConfigs.Configs[i].Path)
			fmt.Println("FileExtension: " + headerConfigs.Configs[i].FileExtension)
			for j := 0; j < len(headerConfigs.Configs); j++ {
				headerRule := headerConfigs.Configs[i].Headers[j]
				fmt.Println(headerRule.Key, ":", headerRule.Value)
			}
			fmt.Println("==============================================================")
		}

		jsonFile.Close()

		return headerConfigs
	} else {
		return nil
	}
}

func main() {

	flag.Parse()

	headerConfigs := initHeaderConfig()

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

	handler := handleReq(http.FileServer(fileSystem), headerConfigs)

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
