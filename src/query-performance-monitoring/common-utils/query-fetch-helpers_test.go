package commonutils_test

import (
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"testing"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
)

func TestFetchVersionSpecificSlowQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.SlowQueriesForV12, false},
		{commonutils.PostgresVersion13, queries.SlowQueriesForV13AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	for _, test := range tests {
		result, err := commonutils.FetchVersionSpecificSlowQueries(test.version)
		if test.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

func TestFetchVersionSpecificBlockingQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.BlockingQueriesForV12AndV13, false},
		{commonutils.PostgresVersion13, queries.BlockingQueriesForV12AndV13, false},
		{commonutils.PostgresVersion14, queries.BlockingQueriesForV14AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	for _, test := range tests {
		result, err := commonutils.FetchVersionSpecificBlockingQueries(test.version)
		if test.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

func TestFetchVersionSpecificIndividualQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.IndividualQuerySearchV12, false},
		{commonutils.PostgresVersion13, queries.IndividualQuerySearchV13AndAbove, false},
		{commonutils.PostgresVersion14, queries.IndividualQuerySearchV13AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	for _, test := range tests {
		result, err := commonutils.FetchVersionSpecificIndividualQueries(test.version)
		if test.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}
