package models

import (
	"golang.org/x/crypto/bcrypt"
	"reflect"
)

type User struct {
	ID       int64
	Name     string
	Password string
	Active   bool
	Friends  map[int]bool
}

func (u *User) IsFriend(otherUser *User) bool {
	_, ok := u.Friends[int(otherUser.ID)]
	return ok
}

func (u *User) SetEncryptedPassword(pw string) {
	b, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	//return string(b)
	u.Password = string(b)
}

func (u *User) SetFriendsMap(friends []int) {
	m := make(map[int]bool)

	for _, friend := range friends {
		m[friend] = true
	}
	u.Friends = m
}

func (u *User) Equals(o *User) bool {
	return u.ID == o.ID && u.Name == o.Name &&
		u.Active == o.Active && u.Password == o.Password &&
		reflect.DeepEqual(u.Friends, o.Friends)
}
