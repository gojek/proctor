package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gojektech/proctor/config"
	"github.com/stretchr/testify/assert"
)

func TestProctorURL(t *testing.T) {
	proctorConfigFileExistedBeforeTest := true

	home := os.Getenv("HOME")
	proctorConfigDir := home + "/.proctor"
	proctorConfigFilePath := proctorConfigDir + "/proctor.yaml"
	existingConfigFileData, err := ioutil.ReadFile(proctorConfigFilePath)
	if err != nil {
		proctorConfigFileExistedBeforeTest = false
		os.Mkdir(proctorConfigDir, os.ModePerm)
	}

	proctorUrl := []byte("PROCTOR_URL: any-random-url.com")
	err = ioutil.WriteFile(proctorConfigFilePath, proctorUrl, 0644)
	assert.NoError(t, err)

	config.InitConfig()
	configuredProctorURL := config.ProctorURL()

	assert.Equal(t, "any-random-url.com", configuredProctorURL)

	if proctorConfigFileExistedBeforeTest {
		err = ioutil.WriteFile(proctorConfigFilePath, existingConfigFileData, 0644)
		assert.NoError(t, err)
	} else {
		err = os.Remove(proctorConfigFilePath)
		assert.NoError(t, err)

		err = os.Remove(proctorConfigDir)
		assert.NoError(t, err)
	}
}
