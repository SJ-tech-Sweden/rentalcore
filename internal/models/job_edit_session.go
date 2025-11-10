package models

import "time"

// JobEditSession represents an active editing session for a job.
type JobEditSession struct {
	SessionID   uint      `json:"session_id" gorm:"primaryKey;column:session_id"`
	JobID       uint      `json:"job_id" gorm:"column:job_id"`
	UserID      uint      `json:"user_id" gorm:"column:user_id"`
	Username    string    `json:"username" gorm:"column:username"`
	DisplayName string    `json:"display_name" gorm:"column:display_name"`
	StartedAt   time.Time `json:"started_at" gorm:"column:started_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
	LastSeen    time.Time `json:"last_seen" gorm:"column:last_seen"`
}

func (JobEditSession) TableName() string {
	return "job_edit_sessions"
}
