# stage 0
FROM --platform=$BUILDPLATFORM golang:latest as builder

ARG TARGETPLATFORM

WORKDIR /go/src/github.com/PierreZ/goStatic
COPY . .

RUN mkdir ./bin && \
    # getting right vars from docker buildx
    # especially to handle linux/arm/v6 for example
    GOOS=$(echo $TARGETPLATFORM | cut -f1 -d/) && \
    GOARCH=$(echo $TARGETPLATFORM | cut -f2 -d/) && \
    GOARM=$(echo $TARGETPLATFORM | cut -f3 -d/ | sed "s/v//" ) && \
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${BUILD_ARGS} -ldflags="-s" -tags netgo -installsuffix netgo -o ./bin/goStatic && \
    mkdir ./bin/etc && \
    ID=$(shuf -i 100-9999 -n 1) && \
    echo $ID && \
    echo "appuser:x:$ID:$ID::/sbin/nologin:/bin/false" > ./bin/etc/passwd && \
    echo "appgroup:x:$ID:appuser" > ./bin/etc/group

# stage 1
FROM scratch
WORKDIR /
COPY --from=builder /go/src/github.com/PierreZ/goStatic/bin/ .
USER appuser
ENTRYPOINT ["/goStatic"]
 
