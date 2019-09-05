package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/plugin"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	securityService SecurityService
	email string
	token string
	proctorConfig config.ProctorConfig
}

func (context *testContext) setUp(t *testing.T) {
	value, available := os.LookupEnv("ENABLE_PLUGIN_INTEGRATION_TEST")
	if available != true || value != "true" {
		t.SkipNow()
	}
	context.proctorConfig = config.Config()
	authPluginBinary := context.proctorConfig.AuthPluginBinary
	authPluginExported := context.proctorConfig.AuthPluginExported
	context.securityService = NewSecurityService(authPluginBinary, authPluginExported, plugin.NewGoPlugin())
	assert.NotNil(t, context.securityService)
	email, _ := os.LookupEnv("GATE_PLUGIN_EMAIL")
	assert.NotEmpty(t, email)
	context.email = email
	token, _ := os.LookupEnv("GATE_PLUGIN_TOKEN")
	assert.NotEmpty(t, token)
	context.token = token
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

func TestSecurityService_AuthSuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := ctx.instance().email
	token := ctx.instance().token

	userDetail, err := ctx.instance().securityService.Auth(email, token)

	assert.NoError(t, err)
	assert.NotNil(t, userDetail)
	assert.NotEmpty(t, userDetail.Email)
	assert.NotEmpty(t, userDetail.Name)
	assert.NotEmpty(t, userDetail.Active)
	assert.NotNil(t, userDetail.Groups)
}

func TestSecurityService_AuthPluginLoadFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := ctx.instance().email
	token := ctx.instance().token
	authPluginExported := ctx.instance().proctorConfig.AuthPluginExported
	service := NewSecurityService("non-existent-plugin", authPluginExported, plugin.NewGoPlugin())

	userDetail, err := service.Auth(email, token)
	assert.EqualError(t, err, fmt.Sprintf("failed to load and instantiate *auth.Auth from plugin location %s and exported name %s", "non-existent-plugin", authPluginExported))
	assert.Nil(t, userDetail)
}

func TestSecurityService_AuthPluginFailedToCast(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := ctx.instance().email
	token := ctx.instance().token

	authPluginBinary := ctx.instance().proctorConfig.AuthPluginBinary
	service := NewSecurityService(authPluginBinary, "Verify", plugin.NewGoPlugin())

	userDetail, err := service.Auth(email, token)
	assert.EqualError(t, err, fmt.Sprintf("failed to load and instantiate *auth.Auth from plugin location %s and exported name %s", authPluginBinary, "Verify"))
	assert.Nil(t, userDetail)
}

func TestSecurityService_VerifySuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := ctx.instance().email
	token := ctx.instance().token

	userDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.NotNil(t, userDetail)

	group := userDetail.Groups
	verified, err := ctx.instance().securityService.Verify(*userDetail, group)
	assert.True(t, verified)
	assert.NoError(t, err)
}

func TestSecurityService_VerifyFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := ctx.instance().email
	token := ctx.instance().token

	userDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.NotNil(t, userDetail)

	group := []string{"non-existent-group", "proctor_executor"}
	verified, err := ctx.instance().securityService.Verify(*userDetail, group)
	assert.False(t, verified)
	assert.NoError(t, err)
}
