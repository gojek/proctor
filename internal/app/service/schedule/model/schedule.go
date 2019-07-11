package model

import (
	dbTypes "proctor/internal/app/service/infra/db/types"
	"time"
)

type Schedule struct {
	ID                 uint64            `db:"id"`
	JobName            string            `db:"job_name"`
	Args               dbTypes.Base64Map `db:"args"`
	Tags               string            `db:"tags"`
	Cron               string            `db:"cron"`
	NotificationEmails string            `db:"notification_emails"`
	UserEmail          string            `db:"user_email"`
	Group              string            `db:"group"`
	Enabled            bool              `db:"enabled"`
	CreatedAt          time.Time         `db:"created_at"`
	UpdatedAt          time.Time         `db:"updated_at"`
}
