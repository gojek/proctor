package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/execution/handler/parameter"
	"proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/execution/status"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/plugin"
	"proctor/pkg/notification/event"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	notificationService NotificationService
	config              config.ProctorConfig
}

func (context *testContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_PLUGIN_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	context.config = config.Config()
	pluginsBinary := context.config.NotificationPluginBinary
	pluginsExportedName := context.config.NotificationPluginExported
	assert.NotEmpty(t, pluginsBinary)
	assert.NotEmpty(t, pluginsExportedName)
	context.notificationService = NewNotificationService(pluginsBinary, pluginsExportedName, plugin.NewGoPlugin())
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	return &testContext{}
}

func TestNotificationService_NotifySuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	executionContextId := uint64(1)
	userEmail := "mrproctor@example.com"
	job := parameter.Job{
		Name: "notification-service-integration-test",
		Args: map[string]string{"data": "dummy"},
	}
	executionContext := model.ExecutionContext{
		ExecutionID: executionContextId,
		UserEmail:   userEmail,
		JobName:     job.Name,
		ImageTag:    "test",
		Args:        job.Args,
		CreatedAt:   time.Now(),
		Status:      status.Finished,
		Output:      types.GzippedText("test"),
	}
	executionEvent := event.NewExecutionEvent(userEmail, executionContext)

	ctx.instance().notificationService.Notify(executionEvent)
}
