package audit

import (
	"testing"

	"proctor/proctord/kubernetes"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
	"proctor/proctord/utility"
)

func TestJobsExecutionAuditing(t *testing.T) {
	mockStore := &storage.MockStore{}
	mockKubeClient := &kubernetes.MockClient{}
	testAuditor := New(mockStore, mockKubeClient)
	jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{
		JobName: "any-job-name",
	}

	mockStore.On("AuditJobsExecution", jobsExecutionAuditLog).Return(nil).Once()

	testAuditor.JobsExecution(jobsExecutionAuditLog)

	mockStore.AssertExpectations(t)
	mockKubeClient.AssertExpectations(t)
}

func TestAuditJobsExecutionStatusAuditing(t *testing.T) {
	mockStore := &storage.MockStore{}
	mockKubeClient := &kubernetes.MockClient{}
	testAuditor := New(mockStore, mockKubeClient)

	jobExecutionID := "job-execution-id"
	jobExecutionStatus := "job-execution-status"

	mockKubeClient.On("JobExecutionStatus", jobExecutionID).Return(jobExecutionStatus, nil)
	mockStore.On("UpdateJobsExecutionAuditLog", jobExecutionID, jobExecutionStatus).Return(nil).Once()

	testAuditor.JobsExecutionStatus(jobExecutionID)

	mockStore.AssertExpectations(t)
	mockKubeClient.AssertExpectations(t)
}

func TestAuditJobsExecutionAndStatusAuditing(t *testing.T) {
	mockStore := &storage.MockStore{}
	mockKubeClient := &kubernetes.MockClient{}
	testAuditor := New(mockStore, mockKubeClient)

	jobExecutionID := "job-execution-id"
	jobExecutionStatus := "job-execution-status"
	jobsExecutionAuditLog := &postgres.JobsExecutionAuditLog{
		JobName:             "any-job-name",
		ExecutionID:         postgres.StringToSQLString(jobExecutionID),
		JobSubmissionStatus: utility.JobSubmissionSuccess,
	}

	mockStore.On("AuditJobsExecution", jobsExecutionAuditLog).Return(nil).Once()

	mockKubeClient.On("JobExecutionStatus", jobExecutionID).Return(jobExecutionStatus, nil)
	mockStore.On("UpdateJobsExecutionAuditLog", jobExecutionID, jobExecutionStatus).Return(nil).Once()

	testAuditor.JobsExecutionAndStatus(jobsExecutionAuditLog)

	mockStore.AssertExpectations(t)
	mockKubeClient.AssertExpectations(t)
}
