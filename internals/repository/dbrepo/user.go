package dbrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/stackninja.pro/goth/internals/models"
	"golang.org/x/crypto/bcrypt"
)

// GetAllUsers retrieves all the user from the database
func (m *neonDBRepo) GetAllUsers() ([]models.User, error) {
	var users []models.User
	rows, err := m.DB.Query(context.Background(), "SELECT id, email, name, category, dob, bio, avatar, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Category, &user.DOB, &user.Bio, &user.Avatar, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUserByID retrieves a user by their ID
func (m *neonDBRepo) GetUserByID(id string) (*models.User, error) {
	var user models.User
	row := m.DB.QueryRow(context.Background(), "SELECT id, email, name, category, dob, bio, avatar, created_at, updated_at FROM users WHERE id = $1", id)
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.Category, &user.DOB, &user.Bio, &user.Avatar, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	//Format the date using a friendly format
	user.DOBFormatted = user.DOB.Format("January 2, 2006")

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (m *neonDBRepo) GetUserByEmail(email string) (*models.User, error) {
    user := &models.User{}
    query := `
        SELECT id, email, password, name, category, dob, bio, avatar, created_at, updated_at
        FROM users
        WHERE email = $1
    `
    row := m.DB.QueryRow(context.Background(), query, email)

    err := row.Scan(
        &user.ID,
        &user.Email,
        &user.Password, // ✅ must scan this
        &user.Name,
        &user.Category,
        &user.DOB,
        &user.Bio,
        &user.Avatar,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return user, nil
}


// CreateUser creates a new user in the database
func (m *neonDBRepo) CreateUser(user models.User) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	// convert id to string and set it on the user
	user.ID = id.String()

	stm, err := m.DB.Prepare(context.Background(), "insert_user", "INSERT INTO users (id, email, password, name, category, dob, bio, avatar, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(context.Background(), stm.Name, user.ID, user.Email, user.Password, user.Name, user.Category, user.DOB, user.Bio, user.Avatar, user.CreatedAt, user.UpdatedAt)

	return err
}

func (m *neonDBRepo) UpdateUser(id string, user models.User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, dob = $3, bio = $4, category = $5, updated_at = NOW()
		WHERE id = $6
	`
	stm, err := m.DB.Prepare(context.Background(), "update_user", query)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(context.Background(), stm.Name, user.Name, user.Email, user.DOB, user.Bio, user.Category, id)

	return err
}

func (m *neonDBRepo) UpdateUserAvatar(userID, filePath string) error {
	stm, err := m.DB.Prepare(context.Background(), "update_user_avatar", 
		"UPDATE users SET avatar = $2 WHERE id = $1")
	if err != nil {
		return err
	}

	// ✅ Correct order: userID first, filePath second
	_, err = m.DB.Exec(context.Background(), stm.Name, userID, filePath)
	return err
}


func (m *neonDBRepo) DeleteUser(id string) error {
	stm, err := m.DB.Prepare(context.Background(), "delete_user", "DELETE FROM users WHERE id = $1")
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(context.Background(), stm.Name, id)

	return err
}

func (m *neonDBRepo) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	row := m.DB.QueryRow(context.Background(), "SELECT id, email, password, name, category, dob, bio, avatar, created_at, updated_at FROM users WHERE email = $1", email)
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Category, &user.DOB, &user.Bio, &user.Avatar, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	// Check if the provided password matches the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}
