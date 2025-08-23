package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/sessions"
	"github.com/stackninja.pro/goth/internals/driver"
	"github.com/stackninja.pro/goth/internals/handlers"
	"github.com/stackninja.pro/goth/src/config"
)



func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if conn, err := Run(ctx); err != nil {
		log.Fatal(err)
	} else {
		defer conn.Conn.Close(context.Background())
	}

}

func Run(ctx context.Context) (*driver.DB, error) {
	var app config.AppConfig


	app.Session = sessions.NewCookieStore([]byte("usermanagementsecret"))

	// Sessions
	app.Session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 3,
		HttpOnly: true,
		Secure:   true,
	}

	

	// DB connection
	log.Println("🔗 Connecting to database...")
	conn, err := driver.ConnectToDB("postgresql://neondb_owner:npg_egURTqLtd89F@ep-divine-cherry-afa5sm0c-pooler.c-2.us-west-2.aws.neon.tech/neondb?sslmode=require&channel_binding=require")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}


	repo := handlers.NewRepository(&app, conn)
	handlers.NewHandlers(repo)

	// Fixed port
	addr := ":8000"
	srv := &http.Server{
		Addr:    addr,
		Handler: Route(),
	}

	// Run server in goroutine
	go func() {
		log.Println("⇨ http://localhost" + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server on %s: %v", addr, err)
		}
	}()

	// Wait for interrupt (Ctrl+C)
	<-ctx.Done()
	log.Println("⏳ Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
	return conn, nil
}
