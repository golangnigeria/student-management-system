package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/stackninja.pro/goth/internals/driver"
	"github.com/stackninja.pro/goth/internals/models"
	"github.com/stackninja.pro/goth/internals/repository"
	"github.com/stackninja.pro/goth/internals/repository/dbrepo"
	"github.com/stackninja.pro/goth/src/config"
	"github.com/stackninja.pro/goth/web/templates"
	"github.com/stackninja.pro/goth/web/templates/components"
	"golang.org/x/crypto/bcrypt"
)

// Repo is the global repository for the application
var Repo *Repository

// Repository holds the application data
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepository creates a new Repository
func NewRepository(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(a, db.Conn),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) HomePage(w http.ResponseWriter, r *http.Request) {
	log.Println("➡️ HomePage handler called")

	users := []models.User{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Avatar: "/static/avatars/1.jpg", Category: 1, CreatedAt: time.Now()},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com", Avatar: "/static/avatars/2.jpg", Category: 2, CreatedAt: time.Now()},
	}

	// Check user session
	user, _, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	log.Println("✅ Logged in user passed to template:", user.Email)

	userMap := map[string]interface{}{
		"users":       users,
		"userSession": user,
	}

	err = templates.HomePage(&models.TemplateData{
		Data: userMap,
	}).Render(r.Context(), w)
	if err != nil {
		log.Println("❌ Template render error:", err)
	}
}

func (m *Repository) AboutPage(w http.ResponseWriter, r *http.Request) {
	// check if the user is authenticated
	user, _, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	// Render the About page
	templates.AboutPage(&models.TemplateData{Data: map[string]interface{}{"title": "Goth Stack Demo", "userSession": user}}).Render(r.Context(), w)
}

func (m *Repository) ShowRegisterPage(w http.ResponseWriter, r *http.Request) {
	// empty user + no errors
	user := &models.User{}

	// Prepare a template
	td := &models.TemplateData{
		Data: map[string]any{
			"user": user,
		},
		StringMap: map[string]string{
			"title": "Create an Account",
		},
	}

	registrationPage := templates.RegistrationPage(td)
	registrationPage.Render(r.Context(), w)
}

func (m *Repository) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var errorMessages []string

	// Parse form
	if err := r.ParseForm(); err != nil {
		errorMessages = append(errorMessages, "Failed to parse form data")
	}

	user.Name = r.FormValue("name")
	user.Email = r.FormValue("email")
	user.Password = r.FormValue("password")
	user.Category, _ = strconv.Atoi(r.FormValue("category"))

	// Validation
	if user.Name == "" {
		errorMessages = append(errorMessages, "Name is required")
	}
	if user.Email == "" {
		errorMessages = append(errorMessages, "Email is required")
	}
	if user.Password == "" {
		errorMessages = append(errorMessages, "Password is required")
	}
	if user.Category == 0 {
		errorMessages = append(errorMessages, "Category is required")
	}

	// Prepare template data
	td := &models.TemplateData{
		Data: map[string]interface{}{
			"user": &user,
		},
		StringMap: map[string]string{
			"title": "Create an Account",
		},

		Errors: errorMessages,
	}

	registrationPage := templates.RegistrationPage(td)

	// If there are validation errors → render error fragment
	if len(errorMessages) > 0 {
		templ.Handler(registrationPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		td.Errors = append(td.Errors, "Failed to hash password")
		templ.Handler(registrationPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}
	user.Password = string(hashedPassword)

	// Defaults
	user.DOB = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	user.Bio = "This is a sample bio."
	user.Avatar = "/static/avatars/default.jpg"

	// Save user
	if err := m.DB.CreateUser(user); err != nil {
		td.Errors = append(td.Errors, "Failed to create user")
		templ.Handler(registrationPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// ✅ Success: redirect via HTMX
	w.Header().Set("HX-Location", "/login")
	w.WriteHeader(http.StatusNoContent)
}

func (m *Repository) LoginPage(w http.ResponseWriter, r *http.Request) {
	templates.LoginPage(&models.TemplateData{}).Render(r.Context(), w)
}

func (m *Repository) LoginUser(w http.ResponseWriter, r *http.Request) {
	var errorMessages []string

	// Parse form
	if err := r.ParseForm(); err != nil {
		errorMessages = append(errorMessages, "Failed to parse form data")
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validation
	if email == "" {
		errorMessages = append(errorMessages, "Email is required")
	}
	if password == "" {
		errorMessages = append(errorMessages, "Password is required")
	}

	// Prepare template data
	td := &models.TemplateData{

		StringMap: map[string]string{
			"title": "Login",
		},

		Errors: errorMessages,
	}

	loginPage := templates.LoginPage(td)

	// If there are validation errors → render error fragment
	if len(errorMessages) > 0 {
		templ.Handler(loginPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Retrieve user by email
	user, err := m.DB.GetUserByEmail(email)
	if err != nil {
		td.Errors = append(td.Errors, "Invalid email or password")
		templ.Handler(loginPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Check password and compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		td.Errors = append(td.Errors, "Invalid email or password")
		templ.Handler(loginPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Create a session and authenticate user
	session, err := m.App.Session.Get(r, "logged-in-user")
	if err != nil {
		td.Errors = append(td.Errors, "Failed to create session")
		templ.Handler(loginPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		td.Errors = append(td.Errors, "Failed to save session")
		templ.Handler(loginPage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// ✅ Success: redirect via HTMX to home page
	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusNoContent)
}

func (m *Repository) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := m.App.Session.Get(r, "logged-in-user")
	if err != nil {
		log.Println("❌ Failed to get session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// remove the user from the session
	delete(session.Values, "user_id")

	// save the change to database
	if err := session.Save(r, w); err != nil {
		log.Println("❌ Failed to save session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Clear session
	session.Options.MaxAge = -1
	session.Values["user_id"] = nil

	//redirect
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (m *Repository) EditProfilePage(w http.ResponseWriter, r *http.Request) {
	// check if the user is authenticated
	user, _, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	// Render the Edit Profile page
	templates.EditProfile(&models.TemplateData{Data: map[string]interface{}{"title": "Goth Stack Demo", "userSession": user}}).Render(r.Context(), w)
}

func (m *Repository) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// check if the user is authenticated
	currentUserProfile, userID, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	var errorMessages []string

	// Parse form
	if err := r.ParseForm(); err != nil {
		errorMessages = append(errorMessages, "Failed to parse form data")
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	dobStr := r.FormValue("dob")
	bio := r.FormValue("bio")

	// Validation
	if name == "" {
		errorMessages = append(errorMessages, "Name is required")
	}
	if email == "" {
		errorMessages = append(errorMessages, "Email is required")
	}
	if dobStr == "" {
		errorMessages = append(errorMessages, "Date of Birth is required")
	}

	// Prepare template data
	td := &models.TemplateData{
		Data: map[string]interface{}{
			"userSession": currentUserProfile,
		},
		StringMap: map[string]string{
			"title": "Edit Profile",
		},

		Errors: errorMessages,
	}

	editProfilePage := components.EditProfileForm(td)

	// If there are validation errors → render error fragment
	if len(errorMessages) > 0 {
		templ.Handler(editProfilePage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Parse DOB
	dob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		td.Errors = append(td.Errors, "Invalid Date of Birth format")
		templ.Handler(editProfilePage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// Update user fields
	user := &models.User{
		ID:       userID,
		Name:     name,
		Email:    email,
		DOB:      dob,
		Bio:      bio,
		Category: currentUserProfile.Category,
	}

	// Update user in DB
	if err := m.DB.UpdateUser(userID, *user); err != nil {
		log.Println("❌ DB update error:", err) // <-- log actual error
		td.Errors = append(td.Errors, "Failed to update profile")
		templ.Handler(editProfilePage, templ.WithFragments("error-messages")).ServeHTTP(w, r)
		return
	}

	// ✅ Success: redirect via HTMX to home page
	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusNoContent)
}

func (m *Repository) ChangeAvatarPage(w http.ResponseWriter, r *http.Request) {
	// check if the user is authenticated
	user, _, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	// Render the Change Avatar page
	templates.ChangeAvatar(&models.TemplateData{Data: map[string]interface{}{"title": "Change Avatar", "userSession": user}}).Render(r.Context(), w)
}

func (m *Repository) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	user, _, err := m.CheckUserAuthentication(w, r)
	if err != nil {
		log.Println("⛔ User not authenticated → redirect triggered")
		return
	}

	var errorMessages []string

	// parse multipart form (10MB max)
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println("❌ Failed to parse multipart form:", err)
		errorMessages = append(errorMessages, "Failed to parse form data")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}

	file, handler, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			errorMessages = append(errorMessages, "No file uploaded")
		} else {
			errorMessages = append(errorMessages, "Failed to retrieve file")
		}
		if len(errorMessages) > 0 {
			templates.ChangeAvatar(&models.TemplateData{
				Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
				Errors: errorMessages,
			}).Render(r.Context(), w)
			return
		}
	}
	defer file.Close()

	// Generate a unique filename
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Println("❌ Failed to generate UUID:", err)
		errorMessages = append(errorMessages, "Failed to generate unique filename")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}
	filename := uuid.String() + filepath.Ext(handler.Filename)

	// ✅ Define upload directory
	uploadDir := filepath.Join(".", "uploads") // ./uploads
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Println("❌ Failed to create uploads dir:", err)
		errorMessages = append(errorMessages, "Failed to prepare upload folder")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}

	// ✅ Save uploaded file inside ./uploads
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Println("❌ Failed to create file:", err)
		errorMessages = append(errorMessages, "Failed to save uploaded file")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Println("❌ Failed to save uploaded file:", err)
		errorMessages = append(errorMessages, "Failed to save uploaded file")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}

	// Update user avatar in DB
	if err := m.DB.UpdateUserAvatar(user.ID, filename); err != nil {
		log.Println("❌ Failed to update user avatar in DB:", err)
		errorMessages = append(errorMessages, "Failed to update user avatar")
		templates.ChangeAvatar(&models.TemplateData{
			Data:   map[string]interface{}{"title": "Change Avatar", "userSession": user},
			Errors: errorMessages,
		}).Render(r.Context(), w)
		return
	}

	// Delete old avatar file if different
	if user.Avatar != "" && user.Avatar != filename {
		oldAvatarPath := filepath.Join(uploadDir, user.Avatar)
		if err := os.Remove(oldAvatarPath); err != nil {
			log.Printf("⚠️ Failed to delete old avatar: %s\n", err)
		}
	}

	// ✅ Success: redirect to profile page
	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusNoContent)
}


func (m *Repository) CheckUserAuthentication(w http.ResponseWriter, r *http.Request) (*models.User, string, error) {
	session, err := m.App.Session.Get(r, "logged-in-user")
	if err != nil {
		log.Println("❌ Failed to get session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, "", err
	}

	userID, ok := session.Values["user_id"]
	if !ok {
		log.Println("⚠️ No user_id in session → redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, "", fmt.Errorf("user not authenticated")
	}

	log.Println("✅ Found user_id in session:", userID)

	user, err := m.DB.GetUserByID(userID.(string))
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("⚠️ No user found in DB for ID:", userID)

			// clear session
			session.Options.MaxAge = -1
			session.Values["user_id"] = nil
			_ = session.Save(r, w)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return nil, "", fmt.Errorf("user not found")
		}

		log.Println("❌ Error fetching user from DB:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, "", err
	}

	log.Println("✅ Authenticated user:", user.Email)
	return user, user.ID, nil
}
