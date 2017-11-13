# stage 0
FROM golang:latest as builder
WORKDIR /go/src/github.com/PierreZ/goStatic
COPY . .

RUN GOARCH=amd64 GOOS=linux go build  -ldflags "-linkmode external -extldflags -static -w"

# stage 1
FROM centurylink/ca-certs
WORKDIR /
COPY --from=builder /go/src/github.com/PierreZ/goStatic/goStatic .
ENTRYPOINT ["/goStatic"]