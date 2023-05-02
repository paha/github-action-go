ARG BUILDPLATFORM=linux/amd64

# Build
FROM --platform=${BUILDPLATFORM} golang:1.20 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${BUILDPLATFORM}

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
    && upx -q -9 /bin/action

# Container
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/action /action

ENTRYPOINT ["/action"]
