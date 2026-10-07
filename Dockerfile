# Build stage
# Pinned to the build platform: Go cross-compiles via the TARGET* args instead
# of running the toolchain under QEMU for arm64. `$BUILDPLATFORM` needs BuildKit.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /major-tag-action ./cmd/main.go

# Runtime stage
FROM alpine:3.24

RUN apk add --no-cache git openssh-client

COPY --from=builder /major-tag-action /usr/local/bin/major-tag-action

ENTRYPOINT ["major-tag-action"]
