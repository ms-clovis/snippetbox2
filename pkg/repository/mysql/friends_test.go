package mysql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ms-clovis/snippetbox/pkg/models"
	"testing"
)

func TestFriendsRepository_FindFriends(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	fr := NewFriendsRepository(db)
	defer fr.CloseDB()

	user := &models.User{
		ID:       1,
		Name:     "foo@test.com",
		Password: "12345678",
		Active:   true,
	}
	rows := sqlmock.NewRows([]string{"watcher"}).
		AddRow(20)

	mock.ExpectQuery("SELECT").WithArgs(user.ID).
		WillReturnRows(rows)

	friends, err := fr.FindFriends(user)

	if err != nil {
		t.Fatal(err)
	}
	if len(friends) < 1 {
		t.Fatal("did not find friends")
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
		t.Error("Not all expectations were met")
	}

}
