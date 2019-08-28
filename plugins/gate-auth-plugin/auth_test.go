package main

import (
	"github.com/stretchr/testify/assert"
	"proctor/pkg/auth"
	"proctor/plugins/gate-auth-plugin/gate"
	"testing"
)

type context interface {
	setUp(t *testing.T)
	tearDown()
	instance() *testContext
}

type testContext struct {
	gateAuth   gateAuth
	gateClient *gate.GateClientMock
}

func (context *testContext) setUp(t *testing.T) {
	context.gateAuth = gateAuth{}
	context.gateClient = &gate.GateClientMock{}
	context.gateAuth.gateClient = context.gateClient
}

func (context *testContext) tearDown() {
}

func (context *testContext) instance() *testContext {
	return context
}

func newContext() context {
	return &testContext{}
}

func TestGateAuth_AuthSuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	email := "w.albertusd@gmail.com"
	token := "unreadabletoken"
	expectedUserDetail := &auth.UserDetail{
		Name:   "William Albertus Dembo",
		Email:  "w.albertusd@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_executor"},
	}
	ctx.instance().gateClient.On("GetUserProfile", email, token).Return(expectedUserDetail, nil)

	actualUserDetail, err := ctx.instance().gateAuth.Auth(email, token)

	assert.NoError(t, err)
	assert.NotNil(t, actualUserDetail)
	assert.Equal(t, expectedUserDetail, actualUserDetail)
}
