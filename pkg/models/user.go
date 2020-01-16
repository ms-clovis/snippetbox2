package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID       int64
	Name     string
	Password string
	Active   bool
	Friends  map[int]bool
}

func (u *User) SetEncryptedPassword(pw string) {
	b, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	//return string(b)
	u.Password = string(b)
}

func (u *User) SetFriendsMap(m map[int]bool) {
	u.Friends = m
}
