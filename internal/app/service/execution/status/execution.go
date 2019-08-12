package status

type ExecutionStatus string

const (
	Received            ExecutionStatus = "RECEIVED"
	RequirementNotMet   ExecutionStatus = "REQUIREMENT_NOT_MET"
	Created             ExecutionStatus = "CREATED"
	CreationFailed      ExecutionStatus = "CREATION_FAILED"
	JobCreationFailed   ExecutionStatus = "JOB_CREATION_FAILED"
	JobReady            ExecutionStatus = "JOB_READY"
	PodCreationFailed   ExecutionStatus = "POD_CREATION_FAILED"
	PodReady            ExecutionStatus = "POD_READY"
	PodFailed           ExecutionStatus = "POD_FAILED"
	FetchPodLogFailed   ExecutionStatus = "FETCH_POD_LOG_FAILED"
	Finished            ExecutionStatus = "FINISHED"
)
