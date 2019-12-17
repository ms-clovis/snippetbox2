package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type Repository interface {
	Fetch(numberToFetch int) ([]*models.Snippet, error)
	FetchAll() ([]*models.Snippet, error)
	GetByID(ID int) (*models.Snippet, error)
	Latest() (*models.Snippet, error)
	Create(m *models.Snippet) (int64, error)
	Delete(m *models.Snippet) (bool, error)
	Update(m *models.Snippet) (bool, error)
}
