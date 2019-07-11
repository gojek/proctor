package status

type ExecutionStatus string

const (
	Received          ExecutionStatus = "RECEIVED"
	RequirementNotMet ExecutionStatus = "REQUIREMENT_NOT_MET"
	Created           ExecutionStatus = "CREATED"
	CreationFailed    ExecutionStatus = "CREATION_FAILED"
)
