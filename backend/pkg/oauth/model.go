package oauthhelper

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"provider"` // "google" or "github"
}

type GetUserInfoReq struct {
	State string
	Code  string
}	
	