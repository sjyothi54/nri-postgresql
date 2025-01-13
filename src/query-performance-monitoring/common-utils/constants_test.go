
package commonutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, 30, MaxQueryThreshold)
	assert.Equal(t, 10, MaxIndividualQueryThreshold)
	assert.Equal(t, 100, PublishThreshold)
	assert.Equal(t, 1000000, RandomIntRange)
	assert.Equal(t, "20060102150405", TimeFormat)
	assert.Equal(t, "PostgreSQL (\\d+)\\.", VersionRegex)
	assert.Equal(t, 12, PostgresVersion12)
	assert.Equal(t, 13, PostgresVersion13)
	assert.Equal(t, 14, PostgresVersion14)
	assert.Equal(t, 2, VersionIndex)
}

func TestErrors(t *testing.T) {
	assert.EqualError(t, ErrParseVersion, "unable to parse PostgreSQL version from string")
	assert.EqualError(t, ErrUnsupportedVersion, "unsupported PostgreSQL version")
	assert.EqualError(t, ErrVersionFetchError, "no rows returned from version query")
	assert.EqualError(t, ErrInvalidModelType, "invalid model type")
}