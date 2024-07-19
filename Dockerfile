# ttynvt builder
FROM alpine:latest AS ttynvt-builder

RUN apk add --no-cache make gcc git libtool autoconf \
            automake pkgconf fuse-dev musl-dev linux-headers
RUN git clone https://gitlab.com/lars-thrane-as/ttynvt.git /work
WORKDIR /work
RUN autoreconf -vif
RUN ./configure
RUN make

# device plugin builder
FROM docker.io/golang:1.21.6-alpine3.19 as device-manager-builder
RUN apk --no-cache add git
RUN mkdir -p /go/src/github.com/yeyus/ttynvt-device-plugin
ADD . /go/src/github.com/yeyus/ttynvt-device-plugin
WORKDIR /go/src/github.com/yeyus/ttynvt-device-plugin
RUN go install \
    -ldflags="-X main.gitDescribe=$(git -C /go/src/github.com/yeyus/ttynvt-device-plugin describe --always --long --dirty)"

# device plugin container
FROM alpine:3.19.1
LABEL \
    org.opencontainers.image.source="https://github.com/yeyus/ttynvt-device-plugin" \
    org.opencontainers.image.authors="Jesus Trujillo <elyeyus@gmail.com>" \
    org.opencontainers.image.licenses="MIT"
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/bin/ttynvt-device-plugin .
CMD ["./ttynvt-device-plugin"]