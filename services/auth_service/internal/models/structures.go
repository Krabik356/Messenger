package models

type RegRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegResp struct {
	Status  string `json:"status"`
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}
