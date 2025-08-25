# -------------------------
# Build Stage
# -------------------------
FROM golang:1.22-alpine AS builder

# Install required tools
RUN apk add --no-cache bash npm git

WORKDIR /app
COPY . .

# Install templ + tailwind
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    npm install -g tailwindcss

# Generate templates, build Tailwind, then compile Go binary
RUN templ generate && \
    tailwindcss -i ./src/app.css -o ./web/static/css/main.css --minify && \
    go build -o bin/goth-demo ./cmd/web

# -------------------------
# Runtime Stage
# -------------------------
FROM alpine:3.18

WORKDIR /app

# Copy only what’s needed
COPY --from=builder /app/bin/goth-demo .
COPY --from=builder /app/web/static ./web/static

EXPOSE 8080

CMD ["./goth-demo"]
