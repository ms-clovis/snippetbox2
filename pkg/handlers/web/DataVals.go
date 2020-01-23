package web

import (
	"github.com/ms-clovis/snippetbox/pkg/models"
)

type DataVals struct {
	Errors          map[string]string
	User            *models.User
	Users           []*models.User
	Snippets        []*models.Snippet
	Snippet         *models.Snippet
	ExpiresDays     string
	Title           string
	Message         string
	IsAuthenticated bool
	CurrentYear     int
	CSRFToken       string
	CurrentLink     string
}
