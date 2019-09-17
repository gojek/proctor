package event

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/status"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() context
}

type executionTestContext struct {
	executionEvent executionEvent
}

func (context *executionTestContext) setUp(t *testing.T) {
}

func (context *executionTestContext) tearDown() {
}

func (context *executionTestContext) instance() context {
	return context
}

func newExecutionTestContext() context {
	ctx := &executionTestContext{}
	return ctx
}

func TestExecutionEvent(t *testing.T) {
	ctx := newExecutionTestContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	userEmail := "mrproctor@example.com"
	executionContextId := uint64(7)
	job := parameter.Job{
		Name: "sample-job-name",
		Args: map[string]string{"argOne": "sample-arg"},
	}
	executionContext := model.ExecutionContext{
		ExecutionID: executionContextId,
		UserEmail:   userEmail,
		JobName:     job.Name,
		ImageTag:    "test",
		Args:        job.Args,
		CreatedAt:   time.Now(),
		Status:      status.Created,
		Output:      types.GzippedText("test"),
	}
	actualEvent := NewExecutionEvent(userEmail, executionContext).(executionEvent)

	expectedEventType := "EXECUTION_EVENT"
	expectedUserData := UserData{
		Email: userEmail,
	}
	jobArgsByte, _ := json.Marshal(job.Args)
	expectedContent := map[string]string{
		"ExecutionID": strconv.FormatUint(executionContext.ExecutionID, 10),
		"JobName":     executionContext.JobName,
		"ImageTag":    executionContext.ImageTag,
		"Args":        string(jobArgsByte),
		"Status":      "CREATED",
	}

	assert.Equal(t, expectedEventType, actualEvent.Type())
	assert.Equal(t, expectedUserData, actualEvent.User())
	assert.Equal(t, expectedContent, actualEvent.Content())
}
