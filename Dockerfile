# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/arch-forge/cli/internal/adapter/cli.Version=${VERSION}" \
    -o /bin/arch_forge ./cmd/archforge

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates git

COPY --from=builder /bin/arch_forge /usr/local/bin/arch_forge

RUN addgroup -S archforge && adduser -S -G archforge archforge
USER archforge

WORKDIR /workspace

ENTRYPOINT ["arch_forge"]
CMD ["--help"]
