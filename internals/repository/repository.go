package repository

import (
	"github.com/stackninja.pro/goth/internals/models"
)

type DatabaseRepo interface {
	GetAllUsers() ([]models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user models.User) error
	UpdateUser(id string, user models.User) error
	UpdateUserAvatar(userID, filePath string) error
	DeleteUser(id string) error
	AuthenticateUser(email, password string) (*models.User, error)
}
