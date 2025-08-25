.PHONY: dev build clean

# Development (watch *.templ + hot-reload Go)
dev:
	@echo "Starting templ + air (hot-reload)…"
	templ generate --watch &
	npx tailwindcss -i ./src/app.css -o ./web/static/css/main.css --watch &
	go run ./cmd/web

# Production build
build:
	templ generate
	npx tailwindcss -i ./src/app.css -o ./web/static/css/main.css --minify
	go build -tags netgo -ldflags "-s -w" -o bin/server ./cmd/web

# Clean artefacts
clean:
	rm -rf tmp bin
