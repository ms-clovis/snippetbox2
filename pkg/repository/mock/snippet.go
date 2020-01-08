package mock

import (
	"database/sql"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"time"
)

var FakeSnippet = &models.Snippet{
	ID:      1,
	Title:   "Fake Snippet",
	Content: "Fake Content",
	Created: time.Now().Add(time.Minute),
	Expires: time.Now().Add(time.Hour), // always expires in the future
	Author:  "mock@test.com",
}

type MockSnippetRepository struct {
	DB *sql.DB
}

func (mr *MockSnippetRepository) Fetch(numberToFetch int) ([]*models.Snippet, error) {
	ret := make([]*models.Snippet, 0)
	ret = append(ret, FakeSnippet)
	return ret, nil
}

func (mr *MockSnippetRepository) FetchAll() ([]*models.Snippet, error) {
	ret := make([]*models.Snippet, 0)
	ret = append(ret, FakeSnippet)
	return ret, nil
}

func (mr *MockSnippetRepository) GetByID(ID int) (*models.Snippet, error) {
	if ID == FakeSnippet.ID {
		return FakeSnippet, nil
	}
	return nil, models.ERRNoRecordFound
}

func (mr *MockSnippetRepository) Latest() (*models.Snippet, error) {
	return FakeSnippet, nil

}

func (mr *MockSnippetRepository) Create(m *models.Snippet) (int64, error) {
	if *m == *FakeSnippet {
		return int64(FakeSnippet.ID), nil
	}
	return -1, models.ERRNoRecordFound
}

// not in usage yet but will be logical delete
func (mr *MockSnippetRepository) Delete(m *models.Snippet) (bool, error) {
	return true, nil
}

// not in usage yet
func (mr *MockSnippetRepository) Update(m *models.Snippet) (bool, error) {
	return true, nil
}

func (mr *MockSnippetRepository) CloseDB() {
	//mr.DB.Close()
	// do nothing
}
