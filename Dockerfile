# stage 0
FROM golang:latest as builder
WORKDIR /go/src/github.com/PierreZ/goStatic
COPY . .
RUN mkdir ./bin && \
    CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -tags netgo -installsuffix netgo -o ./bin/goStatic && \
    mkdir ./bin/etc && \
    cp ./passwd ./bin/etc && \
    cp ./group ./bin/etc

# stage 1
FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/PierreZ/goStatic/bin/ .
USER appuser
ENTRYPOINT ["/goStatic"]
