package mock

import (
	"database/sql"
	"github.com/ms-clovis/snippetbox2/pkg/models"
)

var fakeUser = &models.User{
	ID:       1,
	Name:     "Fake",
	Password: "fake",
	Active:   true,
}

type MockUserRepository struct {
	DB *sql.DB
}

func NewMockUserRepository(DB *sql.DB) *MockUserRepository {
	return &MockUserRepository{DB: DB}
}

func (mr *MockUserRepository) CloseDB() {
	mr.DB.Close()
}

func (mr *MockUserRepository) GetUserByID(id int) (*models.User, error) {
	if id == int(fakeUser.ID) {
		return fakeUser, nil
	}
	return nil, models.ERRNoUserFound
}

func (mr *MockUserRepository) GetUser(name string) (*models.User, error) {
	if name == fakeUser.Name {
		return fakeUser, nil
	}
	return nil, models.ERRNoUserFound
}

func (mr *MockUserRepository) IsAuthenticated(hashedPW string, pw string) bool {
	return true
}
func (mr *MockUserRepository) Create(u *models.User) (int64, error) {
	if u == fakeUser {
		return fakeUser.ID, nil
	}
	return 0, models.ERRUserAlreadyExists
}
func (mr *MockUserRepository) Update(u *models.User) (bool, error) {
	return true, nil
}

func (mr *MockUserRepository) GetUsers(user *models.User) ([]*models.User, error) {
	return []*models.User{fakeUser}, nil
}
