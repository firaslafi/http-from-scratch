# Build Stage
FROM golang:1.27.3-alpine3.23 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
# RUN go mod download

# Copy source code
COPY . .

# Compile static binary
# Docker BuildKit has built-in platform arguments,  injects the target values
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /app/server \
    -trimpath \
    ./cmd/httpserver/
    

# Distroless Runtime Stage
# Use :nonroot to default to an unprivileged user (UID/GID 65532)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /server

# Explicitly ensure execution as the distroless nonroot user
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]