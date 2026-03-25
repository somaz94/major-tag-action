# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /major-tag-action ./cmd/main.go

# Runtime stage
FROM alpine:3.23

RUN apk add --no-cache git openssh-client

COPY --from=builder /major-tag-action /usr/local/bin/major-tag-action

ENTRYPOINT ["major-tag-action"]
