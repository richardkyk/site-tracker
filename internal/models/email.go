package models

type Email struct {
	Email   string `json:"email"`
	URL     string `json:"url"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
