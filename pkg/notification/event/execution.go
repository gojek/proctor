package event

import (
	"encoding/json"
	"strconv"

	"proctor/internal/app/service/execution/model"
)

type executionEvent struct {
	userEmail string
	context   model.ExecutionContext
}

func (evt executionEvent) Type() Type {
	return ExecutionEventType
}

func (evt executionEvent) User() UserData {
	return UserData{
		Email: evt.userEmail,
	}
}

func (evt executionEvent) Content() map[string]string {
	executionContext := evt.context
	jobArgsByte, _ := json.Marshal(executionContext.Args)
	return map[string]string{
		"ExecutionID": strconv.FormatUint(executionContext.ExecutionID, 10),
		"JobName":     executionContext.JobName,
		"ImageTag":    executionContext.ImageTag,
		"Args":        string(jobArgsByte),
		"Status":      string(executionContext.Status),
	}
}

func NewExecutionEvent(userEmail string, context model.ExecutionContext) Event {
	return executionEvent{
		userEmail: userEmail,
		context:   context,
	}
}
