package model

import (
	sqlxTypes "github.com/jmoiron/sqlx/types"
	dbTypes "proctor/internal/app/service/infra/db/types"
	"time"
)

type ExecutionContext struct {
	ExecutionID string                `db:"execution_id"`
	JobName     string                `db:"job_name"`
	UserEmail   string                `db:"user_email"`
	ImageTag    string                `db:"image_tag"`
	Args        dbTypes.Base64Map     `db:"args"`
	Output      sqlxTypes.GzippedText `db:"output"`
	Status      string                `db:"status"`
	CreatedAt   time.Time             `db:"created_at"`
	UpdatedAt   time.Time             `db:"updated_at"`
}
