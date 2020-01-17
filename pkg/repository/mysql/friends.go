package mysql

import (
	"database/sql"
	"errors"
	"github.com/ms-clovis/snippetbox/pkg/models"
)

type FriendsRepository struct {
	DB *sql.DB
}

func NewFriendsRepository(DB *sql.DB) *FriendsRepository {
	return &FriendsRepository{DB: DB}
}

func (fr *FriendsRepository) CloseDB() {
	fr.DB.Close()
}

//noinspection ALL
func (fr *FriendsRepository) FindFriends(user *models.User) ([]int, error) {
	query := "SELECT watched FROM friends WHERE watcher = ?"
	rows, err := fr.DB.Query(query, user.ID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	friends := make([]int, 0)
	var friend int
	for rows.Next() {
		err = rows.Scan(&friend)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	return friends, nil

}

func (fr *FriendsRepository) SetFriend(user *models.User, friendToBe *models.User) (bool, error) {
	insertSQL := "INSERT INTO snippetbox.friends (watcher,watched) VALUES ( ?,? )"
	result, err := fr.DB.Exec(insertSQL, user.ID, friendToBe.ID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, errors.New("did not insert into friends")
	}
	return true, nil

}

func (fr *FriendsRepository) UnFriend(user *models.User, friend *models.User) (bool, error) {
	delSQL := "DELETE FROM snippetbox.friends WHERE watcher = ? AND watched = ?"
	result, err := fr.DB.Exec(delSQL, user.ID, friend.ID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows != 1 {
		return false, errors.New("did not delete from friends")
	}
	return true, nil
}
