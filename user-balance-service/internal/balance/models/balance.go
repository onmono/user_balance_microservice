package models

import (
	"github.com/google/uuid"
	"time"
)

type UserBalance struct {
	ID            uuid.UUID `json:"id,omitempty"`
	UserID        uuid.UUID `json:"user_id"`
	Balance       uint64    `json:"balance,omitempty"`
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
}

type Reserve struct {
	ID            uuid.UUID `json:"id,omitempty"`
	ReserveID     uuid.UUID `json:"reserve_id"`
	UserID        uuid.UUID `json:"user_id"`
	ServiceID     uuid.UUID `json:"service_id"`
	OrderID       uuid.UUID `json:"order_id"`
	Price         uint64    `json:"price"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type AccountingRevenue struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ServiceID uuid.UUID `json:"service_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Sum       uint64    `json:"sum"`
	Timestamp time.Time `json:"timestamp"`
}
