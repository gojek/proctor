package model

import "time"

type ScheduleContext struct {
	ID                 uint64    `json:"id" db:"id"`
	ScheduleId         uint64    `json:"scheduleId" db:"schedule_id"`
	ExecutionContextId uint64    `json:"executionContextId" db:"execution_context_id"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}
