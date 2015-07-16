# goStatic
A really small static web server for Docker

### The goal
My goal is to create to smallest docker container for my web static files. The advantage of Go is that you can generate a fully static binary, so that you don't need anything else.

### Why?
Because the official Golang image is wayyyy to big. (around 1/2Gb as you can see below)

[![](https://badge.imagelayers.io/golang:latest.svg)](https://imagelayers.io/?images=golang:latest 'Get your own badge on imagelayers.io')

For me, the whole point of containers is to have a light container...
Many links should provide you with additionnal info to see my point of view:
 * [Create The Smallest Possible Docker Container](http://blog.xebia.com/2014/07/04/create-the-smallest-possible-docker-container/)
 * [Building Docker Images for Static Go Binaries](https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07)
 * [Small Docker Images For Go Apps](http://www.centurylinklabs.com/small-docker-images-for-go-apps/)

### What are you using?

I'm using [echo](http://echo.labstack.com/) as a  micro web framework because he has great performance, and [golang-builder](https://github.com/CenturyLinkLabs/golang-builder) to generate the static binary.
