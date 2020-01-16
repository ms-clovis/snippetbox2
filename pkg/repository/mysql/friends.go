package mysql

import (
	"database/sql"
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
	query := "SELECT watcher FROM friends WHERE watched = ?"
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
