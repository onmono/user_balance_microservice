package appresponse

import "github.com/google/uuid"

type Message struct {
	Code             int    `json:"code,omitempty"`
	Message          string `json:"message,omitempty"`
	DeveloperMessage string `json:"developer_message,omitempty"`
}

type ResponseDTO struct {
	ID     uuid.UUID `json:"id"`
	Amount float64   `json:"amount"`
}
