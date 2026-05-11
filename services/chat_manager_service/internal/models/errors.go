package models

import "errors"

var InvalidData = errors.New("invalid data")
var ServersError = errors.New("servers error")
var ChatAlreadyExists = errors.New("chat already exists")
var InvalidToken = errors.New("invalid token")
var ExpiredToken = errors.New("token is expired")
var AlreadyExists = errors.New("already exists")
var NoUserInChat = errors.New("no user in chat")
var EmptyChat = errors.New("chat is empty")
var NoOldMessages = errors.New("no another old messages")
