// This small program is just a small web server created in static mode
// in order to provide the smallest docker image possible

package main // import "github.com/PierreZ/goStatic"

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/PierreZ/tlscert"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

var (
	// Def of flags
	portPtr = flag.Int("p", 8043, "The listening port")
	pathPtr = flag.String("static", "/srv/http", "The path for the static files")
	crtPtr  = flag.String("crt", "/etc/ssl/server", "Folder for server.pem and key.pem")
	HTTPPtr = flag.Bool("forceHTTP", false, "Forcing HTTP and not HTTPS")
)

func main() {

	flag.Parse()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())

	// Routes
	e.Static("/", *pathPtr)

	log.Println("Starting goStatic")

	port := ":" + strconv.FormatInt(int64(*portPtr), 10)
	path := *crtPtr

	// Start server with unsecure HTTP
	if *HTTPPtr {
		log.Println("Starting serving", *pathPtr, "on", *portPtr)
		e.Run(port)
		// or with awesome TLS
	} else {
		if _, err := os.Stat(path + "/cert.pem"); os.IsNotExist(err) {
			// Generating certificates
			err := tlscert.GenerateCert(path)
			if err != nil {
				log.Fatalf("Failed generating certs:  %v", err)
			}
		}

		log.Println("Starting serving", *pathPtr, "on", *portPtr)
		e.RunTLS(port, path+"/cert.pem", path+"/key.pem")
	}

}
