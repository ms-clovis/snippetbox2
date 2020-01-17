package repository

import "github.com/ms-clovis/snippetbox/pkg/models"

type FriendsRepository interface {
	CloseDB()
	FindFriends(user *models.User) ([]int, error)
	SetFriend(user *models.User, friendToBe *models.User) (bool, error)
	UnFriend(user *models.User, friend *models.User) (bool, error)
}
