.PHONY: dev build clean

# Development (watch *.templ + hot-reload Go)
dev:
	@echo "Starting templ + air (hot-reload)…"
	templ generate --watch & 
	tailwindcss -i ./src/app.css -o ./web/static/css/main.css
	go run ./cmd/web

# Production build (compile templates + Tailwind, then Go binary)
build:
	templ generate
	tailwindcss -i ./src/app.css -o ./web/static/css/main.css --minify
	go build -o bin/goth-demo ./cmd/web

# Clean artefacts
clean:
	rm -rf tmp bin
