package models

import "time"

type ActivityLog struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Action    string    `db:"action" json:"action"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}
