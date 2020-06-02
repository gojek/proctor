package model

import (
	dbTypes "proctor/internal/app/service/infra/db/types"
	"time"
)

type Schedule struct {
	ID                 uint64            `json:"id" db:"id"`
	JobName            string            `json:"jobName" db:"job_name"`
	Args               dbTypes.Base64Map `json:"args" db:"args"`
	Tags               string            `json:"tags" db:"tags"`
	Cron               string            `json:"cron" db:"cron"`
	NotificationEmails string            `json:"notificationEmails" db:"notification_emails"`
	UserEmail          string            `json:"userEmail" db:"user_email"`
	Group              string            `json:"group" db:"group"`
	Enabled            bool              `json:"enabled" db:"enabled"`
	CreatedAt          time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time         `json:"updatedAt" db:"updated_at"`
}
