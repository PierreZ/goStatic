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
	"github.com/labstack/echo/middleware"
)

var (
	// Def of flags
	portPtr    = flag.Int("p", 8043, "The listening port")
	pathPtr    = flag.String("static", "/srv/http", "The path for the static files")
	crtPtr     = flag.String("crt", "/etc/ssl/server", "Folder for server.pem and key.pem")
	isUnsecure = flag.Bool("forceHTTP", false, "Forcing HTTP and not HTTPS")
)

func main() {

	flag.Parse()

	e := echo.New()

	// Root level middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", *pathPtr) // Serve everything from the current dir
	port := ":" + strconv.FormatInt(int64(*portPtr), 10)
	path := *crtPtr

	// Start server with unsecure HTTP
	if *isUnsecure {
		log.Println("Starting serving", *pathPtr, "on", *portPtr)
		e.Start(port)

	} else { // or with awesome TLS
		if _, err := os.Stat(path + "/cert.pem"); os.IsNotExist(err) {
			// Generating certificates
			err := tlscert.GenerateCert(path)
			if err != nil {
				log.Fatalf("Failed generating certs:  %v", err)
			}
		}

		log.Println("Starting serving", *pathPtr, "on", *portPtr)
		e.StartTLS(port, path+"/cert.pem", path+"/key.pem")
	}
}
