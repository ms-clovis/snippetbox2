package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type UserRepository interface {
	GetUserByID(id int) (*models.User, error)
	GetUser(name string) (*models.User, error)
	IsAuthenticated(u *models.User) (bool, error)
	Create(u *models.User) (int64, error)
}
