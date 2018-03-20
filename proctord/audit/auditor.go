package audit

import (
	"context"

	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
)

type Auditor interface {
	AuditJobsExecution(context.Context)
}

type auditor struct {
	store storage.Store
}

func New(store storage.Store) Auditor {
	return &auditor{
		store: store,
	}
}

func (auditor *auditor) AuditJobsExecution(ctx context.Context) {
	jobSubmissionStatus := ctx.Value(utility.JobSubmissionStatusContextKey).(string)

	if jobSubmissionStatus != utility.JobSubmissionSuccess {
		err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, "", "", "", map[string]string{})
		if err != nil {
			logger.Error("Error auditing jobs execution", err)
		}
		return
	}
	jobName := ctx.Value(utility.JobNameContextKey).(string)
	jobSubmittedForExecution := ctx.Value(utility.JobSubmittedForExecutionContextKey).(string)
	imageName := ctx.Value(utility.ImageNameContextKey).(string)
	jobArgs := ctx.Value(utility.JobArgsContextKey).(map[string]string)

	err := auditor.store.JobsExecutionAuditLog(jobSubmissionStatus, jobName, jobSubmittedForExecution, imageName, jobArgs)
	if err != nil {
		logger.Error("Error auditing jobs execution", err)
	}
}
