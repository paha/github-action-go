# Build
FROM golang:1.17 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apt-get -qq update && \
    apt-get -yqq install upx

WORKDIR /src
COPY . .

RUN go build \
    -a \
    -ldflags "-s -w -extldflags '-static'" \
    -installsuffix cgo \
    -tags netgo \
    -o /bin/action \
    . \
    && strip /bin/action \
    && upx -q -9 /bin/action

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

# Container
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/action /action

USER nobody
ENTRYPOINT ["/action"]
