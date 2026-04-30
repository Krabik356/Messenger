package models

type CreateChatRequest struct {
	UserId   int    `json:"user_id"`
	ChatName string `json:"chat_name"`
}

type CreateChatReturn struct {
	Status string `json:"status"`
}

type AddNewUser struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}
