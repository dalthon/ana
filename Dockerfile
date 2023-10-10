FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

RUN apk update                                     && \
    apk upgrade                                    && \
    apk add --update --no-cache git make           && \
    rm -rf /var/cache/apk/*                        && \
    go install golang.org/x/tools/cmd/godoc@latest && \
    go mod download
