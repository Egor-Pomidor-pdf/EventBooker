package models

import "time"

type Event struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	Capacity  int64     `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

type Booking struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"event_id"`
	UserID    int64     `json:"user_id"`
	Status    string    `json:"status"` // booked | confirmed | cancelled
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
