package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type FriendsRepository interface {
	CloseDB()
	FindFriends(user *models.User) ([]int, error)
}
