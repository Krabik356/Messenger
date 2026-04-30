package models

type CreateChatRequest struct {
	UserId int `json:"user_id"`
}

type CreateChatReturn struct {
	Status string `json:"status"`
}
