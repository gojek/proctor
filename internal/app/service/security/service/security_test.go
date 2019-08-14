package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"proctor/internal/app/service/infra/plugin"
	"proctor/pkg/auth"
	"testing"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	pluginBinary    string
	exportedName    string
	goPlugin        *plugin.GoPluginMock
	securityService SecurityService
	auth            *auth.AuthMock
}

func (context *testContext) setUp(t *testing.T) {
	context.pluginBinary = "plugin-binary"
	assert.NotEmpty(t, context.pluginBinary)
	context.exportedName = "exported-name"
	assert.NotEmpty(t, context.exportedName)
	context.goPlugin = &plugin.GoPluginMock{}
	assert.NotNil(t, context.goPlugin)
	context.securityService = NewSecurityService(context.pluginBinary, context.exportedName, context.goPlugin)
	assert.NotNil(t, context.securityService)
	context.auth = &auth.AuthMock{}
	assert.NotNil(t, context.auth)
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

	email := "jasoet@go-jek.com"
	token := "randomtokenforthesakeofopensource"

	userDetail := &auth.UserDetail{
		Name:   "Deny Prasetyo",
		Email:  "jasoet87@gmail.com",
		Active: true,
		Group:  []string{"system", "proctor_executor"},
	}
	ctx.instance().auth.On("Auth", email, token).Return(userDetail, nil)

	ctx.instance().goPlugin.On("Load", ctx.instance().pluginBinary, ctx.instance().exportedName).Return(ctx.instance().auth, nil)

	expectedUserDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.NoError(t, err)
	assert.NotNil(t, expectedUserDetail)
	assert.Equal(t, userDetail, expectedUserDetail)
}

func TestSecurityService_AuthPluginLoadFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "jasoet@go-jek.com"
	token := "randomtokenforthesakeofopensource"

	ctx.instance().goPlugin.On("Load", ctx.instance().pluginBinary, ctx.instance().exportedName).Return(ctx.instance().auth, fmt.Errorf("load error bro"))

	expectedUserDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.EqualError(t, err, fmt.Sprintf("failed to load and instantiate *auth.Auth from plugin location %s and exported name %s", ctx.instance().pluginBinary, ctx.instance().exportedName))
	assert.Nil(t, expectedUserDetail)
}

func TestSecurityService_AuthPluginFailedToCast(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "jasoet@go-jek.com"
	token := "randomtokenforthesakeofopensource"

	userDetail := &auth.UserDetail{
		Name:   "Deny Prasetyo",
		Email:  "jasoet87@gmail.com",
		Active: true,
		Group:  []string{"system", "proctor_executor"},
	}

	ctx.instance().goPlugin.On("Load", ctx.instance().pluginBinary, ctx.instance().exportedName).Return(userDetail, nil)

	expectedUserDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.EqualError(t, err, fmt.Sprintf("failed to load and instantiate *auth.Auth from plugin location %s and exported name %s", ctx.instance().pluginBinary, ctx.instance().exportedName))
	assert.Nil(t, expectedUserDetail)
}

func TestSecurityService_VerifySuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "jasoet@go-jek.com"
	token := "randomtokenforthesakeofopensource"

	userDetail := &auth.UserDetail{
		Name:   "Deny Prasetyo",
		Email:  "jasoet87@gmail.com",
		Active: true,
		Group:  []string{"system", "proctor_executor"},
	}

	ctx.instance().auth.On("Auth", email, token).Return(userDetail, nil)

	ctx.instance().goPlugin.On("Load", ctx.instance().pluginBinary, ctx.instance().exportedName).Return(ctx.instance().auth, nil)

	expectedUserDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.NoError(t, err)
	assert.NotNil(t, expectedUserDetail)
	assert.Equal(t, userDetail, expectedUserDetail)

	group := []string{"system", "proctor_executor"}
	ctx.instance().auth.On("Verify", *expectedUserDetail, group).Return(true, nil)
	verified, err := ctx.instance().securityService.Verify(*expectedUserDetail, group)
	assert.True(t, verified)
	assert.NoError(t, err)
}

func TestSecurityService_VerifyFailed(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)

	email := "jasoet@go-jek.com"
	token := "randomtokenforthesakeofopensource"

	userDetail := &auth.UserDetail{
		Name:   "Deny Prasetyo",
		Email:  "jasoet87@gmail.com",
		Active: true,
		Group:  []string{"system", "proctor_executor"},
	}

	ctx.instance().auth.On("Auth", email, token).Return(userDetail, nil)

	ctx.instance().goPlugin.On("Load", ctx.instance().pluginBinary, ctx.instance().exportedName).Return(ctx.instance().auth, nil)

	expectedUserDetail, err := ctx.instance().securityService.Auth(email, token)
	assert.NoError(t, err)
	assert.NotNil(t, expectedUserDetail)
	assert.Equal(t, userDetail, expectedUserDetail)

	group := []string{"system", "proctor_executor"}
	ctx.instance().auth.On("Verify", *expectedUserDetail, group).Return(false, fmt.Errorf("verify error"))
	verified, err := ctx.instance().securityService.Verify(*expectedUserDetail, group)
	assert.False(t, verified)
	assert.EqualError(t, err, "verify error")
}


