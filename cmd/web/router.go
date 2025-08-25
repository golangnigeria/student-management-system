package main

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/stackninja.pro/goth/internals/handlers"
)

func Route() chi.Router {
	r := chi.NewRouter()

	// serve CSS and images straight from disk
	workDir, _ := filepath.Abs(".")
	staticDir := http.Dir(filepath.Join(workDir, "web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))

	// in your router setup
	fs := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", fs))

	// pages
	r.Get("/", handlers.Repo.HomePage)
	r.Get("/about", handlers.Repo.AboutPage)
	r.Get("/profile/edit", handlers.Repo.EditProfilePage)
	r.Post("/profile/update", handlers.Repo.UpdateProfile)
	r.Get("/profile/edit/avatar", handlers.Repo.ChangeAvatarPage)
	r.Post("/upload-avatar", handlers.Repo.UploadAvatar)

	// registration routes
	r.Get("/register", handlers.Repo.ShowRegisterPage)
	r.Post("/register", handlers.Repo.RegisterUser)

	// login routes
	r.Get("/login", handlers.Repo.LoginPage)
	r.Post("/login", handlers.Repo.LoginUser)

	// logout route
	r.Get("/logout", handlers.Repo.LogoutUser)

	return r
}
