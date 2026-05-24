package models

import (
	"time"
)

type Subscription struct {
	ID          int       `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"name"`
	Price       int       `json:"price" db:"price"`
	UserID      string    `json:"user_id" db:"user_id"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
}
