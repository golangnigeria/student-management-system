# syntax=docker/dockerfile:1

### Stage 1: Build
FROM golang:1.24 AS builder

# Install Node + Tailwind v3
RUN apt-get update && apt-get install -y nodejs npm && \
    npm install -g tailwindcss@3.4.13 postcss autoprefixer && \
    go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Generate templates
RUN templ generate

# Build Tailwind CSS (minified, v3)
RUN tailwindcss -i ./src/app.css -o ./web/static/css/main.css --minify

# Build Go binary
RUN go build -tags netgo -ldflags="-s -w" -o bin/server ./cmd/web


### Stage 2: Runtime
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary + web assets
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/web /app/web

EXPOSE 8000

CMD ["/app/server"]
