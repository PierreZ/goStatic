# goStatic
A really small static web server for Docker

[![](https://badge.imagelayers.io/pierrezemb/gostatic:latest.svg)](https://imagelayers.io/?images=pierrezemb/gostatic:latest 'Get your own badge on imagelayers.io')
[![GoDoc](https://godoc.org/github.com/PierreZ/goStatic?status.svg)](https://godoc.org/github.com/PierreZ/goStatic)

### The goal
My goal is to create to smallest docker container for my web static files. The advantage of Go is that you can generate a fully static binary, so that you don't need anything else.

### Features
 * A fully static web server in 5MB
 * Web server build for Docker
 * HTTPS by default
 * Can generate certificate on his onw
 * Light container
 * More security than official images (see below)
 * Log enabled

### Why?
Because the official Golang image is wayyyy to big (around 1/2Gb as you can see below) and could be unsecure.

[![](https://badge.imagelayers.io/golang:latest.svg)](https://imagelayers.io/?images=golang:latest 'Get your own badge on imagelayers.io')

For me, the whole point of containers is to have a light container...
Many links should provide you with additionnal info to see my point of view:

 * [Over 30% of Official Images in Docker Hub Contain High Priority Security Vulnerabilities](http://www.banyanops.com/blog/analyzing-docker-hub/)
 * [Create The Smallest Possible Docker Container](http://blog.xebia.com/2014/07/04/create-the-smallest-possible-docker-container/)
 * [Building Docker Images for Static Go Binaries](https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07)
 * [Small Docker Images For Go Apps](http://www.centurylinklabs.com/small-docker-images-for-go-apps/)

### How to use
```
// HTTPS server
docker run -d -p 443:8043 -v path/to/website:/srv/http --name goStatic pierrezemb/gostatic
// HTTP server
docker run -d -p 80:8043 -v path/to/website:/srv/http --name goStatic pierrezemb/gostatic --forceHTTP
```

### Wow, such container! What are you using?

I'm using [echo](http://echo.labstack.com/) as a  micro web framework because he has great performance, and [golang-builder](https://github.com/CenturyLinkLabs/golang-builder) to generate the static binary (command line in the makefile)

I'm also using the centurylink/ca-certs image instead of the scratch image to avoid this error:

```
x509: failed to load system roots and no roots provided
```

The centurylink/ca-certs image is simply the scratch image with the most common root CA certificates pre-installed. The resulting image is only 258 kB which is still a good starting point for creating your own minimal images.
