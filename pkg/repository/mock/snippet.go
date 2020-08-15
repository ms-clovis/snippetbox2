package mock

import (
	"database/sql"
	"github.com/ms-clovis/snippetbox2/pkg/models"
	"time"
)

var FakeSnippet = &models.Snippet{
	ID:      1,
	Title:   "Fake Snippet",
	Content: "Fake Content",
	Created: time.Time{},
	Expires: time.Time{}, // always expires in the future
	Author:  "mock@test.com",
}

type MockSnippetRepository struct {
	DB *sql.DB
}

func (mr *MockSnippetRepository) Fetch(user *models.User, numberToFetch int) ([]*models.Snippet, error) {
	ret := make([]*models.Snippet, 0)
	ret = append(ret, FakeSnippet)
	return ret, nil
}

func (mr *MockSnippetRepository) FetchAll(user *models.User) ([]*models.Snippet, error) {
	ret := make([]*models.Snippet, 0)
	ret = append(ret, FakeSnippet)
	return ret, nil
}

func (mr *MockSnippetRepository) GetByID(user *models.User, ID int) (*models.Snippet, error) {
	if ID == FakeSnippet.ID {
		return FakeSnippet, nil
	}
	return nil, models.ERRNoRecordFound
}

func (mr *MockSnippetRepository) Latest(user *models.User) (*models.Snippet, error) {
	return FakeSnippet, nil

}

func (mr *MockSnippetRepository) Create(user *models.User, m *models.Snippet) (int64, error) {

	return int64(FakeSnippet.ID), nil
}

// not in usage yet but will be logical delete
func (mr *MockSnippetRepository) Delete(user *models.User, m *models.Snippet) (bool, error) {
	return true, nil
}

// not in usage yet
func (mr *MockSnippetRepository) Update(user *models.User, m *models.Snippet) (bool, error) {
	return true, nil
}

func (mr *MockSnippetRepository) CloseDB() {
	//mr.DB.Close()
	// do nothing
}
