package model

import (
	"time"

	sqlxTypes "github.com/jmoiron/sqlx/types"

	"proctor/internal/app/service/execution/status"
	dbTypes "proctor/internal/app/service/infra/db/types"
)

type ExecutionContext struct {
	ExecutionID uint64                 `db:"id"`
	JobName     string                 `db:"job_name"`
	Name        string                 `db:"name"`
	UserEmail   string                 `db:"user_email"`
	ImageTag    string                 `db:"image_tag"`
	Args        dbTypes.Base64Map      `db:"args"`
	Output      sqlxTypes.GzippedText  `db:"output"`
	Status      status.ExecutionStatus `db:"status"`
	CreatedAt   time.Time              `db:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at"`
}
