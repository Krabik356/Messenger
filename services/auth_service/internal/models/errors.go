package models

import "errors"

var InvalidName = errors.New("invalid name")
var InvalidPassword = errors.New("invalid password")
var InvalidEmail = errors.New("invalid email")

var NullName = errors.New("null name")
var NullPassword = errors.New("null password")
var NullEmail = errors.New("null email")

var AlreadyExists = errors.New("already exists")

var ServersError = errors.New("servers error")
