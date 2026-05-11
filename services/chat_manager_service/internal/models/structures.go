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

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Message struct {
	Id   int    `json:"id"`
	User User   `json:"writer"`
	Text string `json:"text"`
}

type MessageSlice struct {
	Messages []Message `json:"messages"`
}

type UsersSlice struct {
	Users []User `json:"users"`
}

type Chat struct {
	Id       int          `json:"id"`
	Name     string       `json:"name"`
	Users    UsersSlice   `json:"users"`
	Messages MessageSlice `json:"messages"`
}

type GetUsersDataResponse struct {
	Chats []Chat `json:"chats"`
}
