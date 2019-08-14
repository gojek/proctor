package plugin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"proctor/internal/app/service/infra/config"
	"testing"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	pluginBinary string
	exportedName string
	goPlugin     GoPlugin
}

func (context *testContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	context.pluginBinary = config.AuthPluginBinary()
	assert.NotEmpty(t, context.pluginBinary)
	context.exportedName = config.AuthPluginExported()
	context.goPlugin = NewGoPlugin()
	assert.NotNil(t, context.goPlugin)
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	ctx := &testContext{}
	return ctx
}

func TestGoPlugin_LoadSuccessfully(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	raw, err := ctx.instance().goPlugin.Load(ctx.instance().pluginBinary, ctx.instance().exportedName)
	assert.NoError(t, err)
	assert.NotNil(t, raw)
}

func TestGoPlugin_LoadPluginFailed(t *testing.T) {
	t.SkipNow()
	ctx := newContext()
	ctx.setUp(t)

	binary := "non-existing-binary"
	raw, err := ctx.instance().goPlugin.Load(binary, ctx.instance().exportedName)
	assert.EqualError(t, err, fmt.Sprintf("failed to load plugin binary from location: %s", binary))
	assert.Nil(t, raw)
}

func TestGoPlugin_LoadExportedFailed(t *testing.T) {
	t.SkipNow()
	ctx := newContext()
	ctx.setUp(t)

	exportedName := "non-existing-exported"
	raw, err := ctx.instance().goPlugin.Load(ctx.instance().pluginBinary, exportedName)
	assert.EqualError(t, err, fmt.Sprintf("failed to Lookup plugin binary from location: %s with Exported Name: %s", ctx.instance().pluginBinary, exportedName))
	assert.Nil(t, raw)
}
