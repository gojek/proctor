package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"proctor/pkg/auth"
	"proctor/plugins/gate-auth-plugin/gate"
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

func TestGateAuth_AuthWrongToken(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	email := "w.albertusd@gmail.com"
	token := "unreadabletoken"
	expectedError := errors.New("authentication failed, please check your access token")
	var userDetail *auth.UserDetail
	ctx.instance().gateClient.On("GetUserProfile", email, token).Return(userDetail, expectedError)

	actualUserDetail, actualError := ctx.instance().gateAuth.Auth(email, token)

	assert.Nil(t, actualUserDetail)
	assert.Error(t, actualError)
	assert.Equal(t, expectedError.Error(), actualError.Error())
}

func TestGateAuth_AuthWrongEmail(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	email := "w.albertusd@gmail.com"
	token := "unreadabletoken"
	expectedError := errors.New("user not found for email w.albertusd@gmail.com")
	var userDetail *auth.UserDetail
	ctx.instance().gateClient.On("GetUserProfile", email, token).Return(userDetail, expectedError)

	actualUserDetail, actualError := ctx.instance().gateAuth.Auth(email, token)

	assert.Nil(t, actualUserDetail)
	assert.Error(t, actualError)
	assert.Equal(t, expectedError.Error(), actualError.Error())
}

func TestGateAuth_VerifySuccess(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	userDetail := auth.UserDetail{
		Name:   "William Albertus Dembo",
		Email:  "w.albertusd@gmail.com",
		Active: true,
		Groups: []string{"system", "proctor_executor"},
	}
	requiredGroups := []string{"system"}

	result, err := ctx.instance().gateAuth.Verify(userDetail, requiredGroups)

	assert.Equal(t, true, result)
	assert.NoError(t, err)
}

func TestGateAuth_VerifyInactiveUser(t *testing.T) {
	ctx := newContext()
	ctx.setUp(t)
	defer ctx.tearDown()

	userDetail := auth.UserDetail{
		Name:   "William Albertus Dembo",
		Email:  "w.albertusd@gmail.com",
		Active: false,
		Groups: []string{"system", "proctor_executor"},
	}
	requiredGroups := []string{"system"}

	result, err := ctx.instance().gateAuth.Verify(userDetail, requiredGroups)

	assert.Equal(t, false, result)
	assert.NoError(t, err)
}
