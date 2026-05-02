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
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SendMessageRequest struct {
	Id      int    `json:"id"`
	ChatId  int    `json:"chatId"`
	Message string `json:"message"`
}

type SendMessageReturn struct {
	Status    string `json:"status"`
	MessageId int    `json:"message_id"`
}
