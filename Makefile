.PHONY: dev build clean

# Development (watch *.templ + hot-reload Go)
dev:
	@echo "Starting templ + air (hot-reload)…"
	templ generate --watch & 
	tailwindcss -i .\src\app.css -o .\web\static\css\main.css --watch
	go run ./cmd/web


# Production build (compile templates, then Go)
build:
	templ generate
	go build -o bin/goth-demo ./cmd/web
	tailwindcss -i .\src\app.css -o .\web\static\css\main.css --watch

# Clean artefacts
clean:
	rm -rf tmp bin