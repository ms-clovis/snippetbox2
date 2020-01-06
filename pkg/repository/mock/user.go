package mock

import (
	"database/sql"
	"github.com/ms-clovis/snippetbox/pkg/models"
)

var fakeUser = &models.User{
	ID:       1,
	Name:     "Fake",
	Password: "fake",
	Active:   true,
}

type MockRepository struct {
	DB *sql.DB
}

func (mr *MockRepository) GetUserByID(id int) (*models.User, error) {
	if id == int(fakeUser.ID) {
		return fakeUser, nil
	}
	return nil, models.ERRNoUserFound
}

func (mr *MockRepository) GetUser(name string) (*models.User, error) {
	if name == fakeUser.Name {
		return fakeUser, nil
	}
	return nil, models.ERRNoUserFound
}

func (mr *MockRepository) IsAuthenticated(hashedPW string, pw string) bool {
	return true
}
func (mr *MockRepository) Create(u *models.User) (int64, error) {
	if *u == *fakeUser {
		return fakeUser.ID, nil
	}
	return 0, models.ERRUserAlreadyExists
}
