package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type UserRepository interface {
	GetUserByID(id int) (*models.User, error)
	GetUser(name string) (*models.User, error)
	IsAuthenticated(hashedPW string, pw string) bool
	Create(u *models.User) (int64, error)
	Update(u *models.User) (bool, error)
}
