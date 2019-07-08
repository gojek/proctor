package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToSQLStringForEmptyString(t *testing.T) {
	sqlString := StringToSQLString("")
	assert.Equal(t, false, sqlString.Valid)
}

func TestStringToSQLStringForNonEmptyString(t *testing.T) {
	sqlString := StringToSQLString("any")

	assert.Equal(t, true, sqlString.Valid)
	assert.Equal(t, "any", sqlString.String)
}
