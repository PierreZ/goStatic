package main // import "github.com/PierreZ/goStatic"

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

func main() {
	// Def & parsing of flags
	portPtr := flag.Int("p", 8080, "The proxy listening port")
	pathPtr := flag.String("path for static file", "/var/www", "The path for the static files")
	crtPtr := flag.String("crt", "/etc/ssl/server", "The path and the name whithout extension for your CRT/key for TLS. Example: /etc/ssl/server for /etc/ssl/server.{crt,key}")
	HTTPPtr := flag.Bool("forceHTTP", false, "Forcing HTTP and not HTTPS")

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

	// Start server with unsecure HTTP
	// or with awesome TLS
	if *HTTPPtr {
		e.Run(port)

	} else {

		if *crtPtr != "/etc/ssl/server" {
			generateCert(*crtPtr)
		}

		e.RunTLS(port, *pathPtr+".pem", *pathPtr+".crt")
	}

}

// Mostly based on https://www.socketloop.com/tutorials/golang-create-x509-certificate-private-and-public-keys
func generateCert(path string) {

	template := &x509.Certificate{
		IsCA: true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"Earth"},
			Organization: []string{"Hooli"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 5, 5),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		log.Println(err)
	}

	// save cert
	ioutil.WriteFile(path+".crt", cert, 0777)
	fmt.Println("certificate saved to", path+".crt")

	// this will create plain text PEM file.
	pemfile, _ := os.Create(path + ".pem")
	var pemkey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privatekey)}
	pem.Encode(pemfile, pemkey)
	pemfile.Close()
}
