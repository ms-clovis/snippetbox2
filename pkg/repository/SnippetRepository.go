package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type SnippetRepository interface {
	Fetch(user *models.User, numberToFetch int) ([]*models.Snippet, error)
	FetchAll(user *models.User) ([]*models.Snippet, error)
	GetByID(user *models.User, ID int) (*models.Snippet, error)
	Latest(user *models.User) (*models.Snippet, error)
	Create(user *models.User, m *models.Snippet) (int64, error)
	Delete(user *models.User, m *models.Snippet) (bool, error)
	Update(user *models.User, m *models.Snippet) (bool, error)
	CloseDB()
}
