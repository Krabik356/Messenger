package models

type RegRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegResp struct {
	Status string `json:"status"`
}

type RegForKafka struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResp struct {
	Status  string `json:"status"`
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}

type RefreshResp struct {
	Status  string `json:"status"`
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}
