package audit

import (
	"context"
	"testing"

	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/utility"
)

func TestExecutionAuditor(t *testing.T) {
	mockStore := &storage.MockStore{}
	testAuditor := New(mockStore)

	jobName := "any-job-name"
	executedJobName := "proctor-123"
	imageName := "any/image:name"
	jobArgs := map[string]string{"key": "value"}

	ctx := context.WithValue(context.Background(), utility.JobSubmissionStatusContextKey, utility.JobSubmissionSuccess)
	ctx = context.WithValue(ctx, utility.JobNameContextKey, jobName)
	ctx = context.WithValue(ctx, utility.JobSubmittedForExecutionContextKey, executedJobName)
	ctx = context.WithValue(ctx, utility.ImageNameContextKey, imageName)
	ctx = context.WithValue(ctx, utility.JobArgsContextKey, jobArgs)

	mockStore.On("JobsExecutionAuditLog", utility.JobSubmissionSuccess, jobName, executedJobName, imageName, jobArgs).Return(nil).Once()

	testAuditor.AuditJobsExecution(ctx)

	mockStore.AssertExpectations(t)
}

func TestExecutionAuditorClientError(t *testing.T) {
	mockStore := &storage.MockStore{}
	testAuditor := New(mockStore)

	ctx := context.WithValue(context.Background(), utility.JobSubmissionStatusContextKey, utility.JobSubmissionClientError)

	mockStore.On("JobsExecutionAuditLog", utility.JobSubmissionClientError, "", "", "", map[string]string{}).Return(nil).Once()

	testAuditor.AuditJobsExecution(ctx)

	mockStore.AssertExpectations(t)
}

func TestExecutionAuditorServerError(t *testing.T) {
	mockStore := &storage.MockStore{}
	testAuditor := New(mockStore)

	ctx := context.WithValue(context.Background(), utility.JobSubmissionStatusContextKey, utility.JobSubmissionServerError)

	mockStore.On("JobsExecutionAuditLog", utility.JobSubmissionServerError, "", "", "", map[string]string{}).Return(nil).Once()

	testAuditor.AuditJobsExecution(ctx)

	mockStore.AssertExpectations(t)
}
