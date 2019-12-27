package models

import "errors"

var ERRNoRecordFound = errors.New("No Matching Record Found")
var ERRUserAlreadyExists = errors.New("User Already Exists")
var ERRMustHaveName = errors.New("Must Have a UserName")
var ERRMustBeValidEmailAddress = errors.New("Username Must Be A Valid Email Address")
var ERRMustHavePassword = errors.New("Password Can Not Be Blank")

var ERRNoUserFound = errors.New("No Matching User")

type Model interface {
	IsModel() bool
}
