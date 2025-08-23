package config

import "github.com/gorilla/sessions"

// AppConfig holds the application configuration
type AppConfig struct {
	Session *sessions.CookieStore
}
