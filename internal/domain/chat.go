package domain

type ChatRequest struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}
