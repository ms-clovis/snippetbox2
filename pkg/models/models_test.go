package models

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestMustBeValidEmailAddress(t *testing.T) {
	text := ERRMustBeValidEmailAddress.Error()
	if text != "Username Must Be A Valid Email Address" {
		t.Error("Error text is not correct")
	}
}

func TestUser_SetEncryptedPassword(t *testing.T) {
	pw := "123456"
	user := &User{}
	user.SetEncryptedPassword(pw)
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pw))

	if err != nil {
		t.Error(err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("123457"))

	if err == nil {
		t.Error("Hash and Password should not be the same")
	}
}
