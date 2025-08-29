package entity

import "time"

type RefreshSession struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
}
