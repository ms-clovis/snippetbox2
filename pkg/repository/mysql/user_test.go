package mysql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ms-clovis/snippetbox2/pkg/models"
	"testing"
)

func TestUserRepository_IsAuthenticated(t *testing.T) {
	password := "123456"
	alphaPW := "abcdef"
	user := &models.User{
		ID:       1,
		Name:     "foo@test.com",
		Password: password,
		Active:   true,
	}

	user.SetEncryptedPassword(password)
	ur := UserRepository{}

	isAuth := ur.IsAuthenticated(user.Password, password)

	if !isAuth {
		t.Error("Did not authenticate")
	}
	user.SetEncryptedPassword(alphaPW)
	isAuth = ur.IsAuthenticated(user.Password, password)

	if isAuth {
		t.Error("Did not authenticate")
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	ur := NewUserRepository(db)
	defer ur.CloseDB()

	user := &models.User{
		ID:       1,
		Name:     "foo@test.com",
		Password: "12345678",
		Active:   true,
	}
	rows := sqlmock.NewRows([]string{"id", "name", "password", "active"}).
		AddRow(user.ID, user.Name, user.Password, user.Active)

	mock.ExpectQuery("SELECT").WithArgs(1).
		WillReturnRows(rows)

	u, err := ur.GetUserByID(int(user.ID))
	if err != nil {
		t.Error(err)
	}
	if user.ID != u.ID || u.Name != user.Name ||
		user.Password != u.Password || user.Active != u.Active {
		t.Error("Did not retrieve user by ID")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
		t.Error("Not all expectations were met")
	}
}

func TestUserRepository_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
	}
	ur := NewUserRepository(db)
	defer ur.CloseDB()
	user := &models.User{
		ID:       1,
		Name:     "foo@test.com",
		Password: "123456",
		Active:   true,
	}
	rows := sqlmock.NewRows([]string{"id", "name", "password", "active"}).
		AddRow(user.ID, user.Name, user.Password, user.Active)
	mock.ExpectQuery("SELECT").WithArgs(user.Name).
		WillReturnRows(rows)

	u, err := ur.GetUser(user.Name)
	if err != nil {
		t.Error(err)
	}

	if !u.Equals(user) {
		t.Error("Did not retrieve correct user by name")
	}
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
	}
	ur := NewUserRepository(db)
	defer ur.CloseDB()

	name := "foo@test.com"
	user := &models.User{

		Name:     name,
		Password: "123456",
		Active:   true,
	}

	user.SetEncryptedPassword(user.Password)

	mock.ExpectExec("INSERT").
		WithArgs(user.Name, user.Password, user.Active).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = ur.Create(user)

	if err != nil {
		t.Error(err)
	}
	if user.Name != name && user.Active != true {
		t.Error()
	}

}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
	}
	ur := NewUserRepository(db)
	defer ur.CloseDB()

	name := "foo@test.com"
	user := &models.User{
		ID:       1,
		Name:     name,
		Password: "123456",
		Active:   true,
	}

	user.SetEncryptedPassword(user.Password)
	mock.ExpectExec("UPDATE").
		WithArgs(user.Name, user.Password, user.Active, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	worked, err := ur.Update(user)
	if !worked || err != nil {
		t.Fatal("Did not update user")
	}

}
