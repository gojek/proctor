package _type

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBase64Map(t *testing.T) {
	maps := Base64Map{
		"name":    "Bimo.zx",
		"address": "Royal Orchid",
		"age":     "17",
	}
	v, err := maps.Value()
	if err != nil {
		t.Errorf("Was not expecting an error")
	}

	expectedMaps := Base64Map{}
	err = (&expectedMaps).Scan(v)
	if err != nil {
		t.Errorf("Was not expecting an error")
	}

	assert.Equal(t, "Bimo.zx", expectedMaps["name"])
	assert.Equal(t, "Royal Orchid", expectedMaps["address"])
	assert.Equal(t, "17", expectedMaps["age"])
}
