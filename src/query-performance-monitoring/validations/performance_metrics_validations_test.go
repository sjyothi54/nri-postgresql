package validations

import (
	"testing"

	performanceDbConnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_isExtensionEnabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	enabled, err := isExtensionEnabled(conn, "pg_stat_statements")
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func Test_CheckSlowQueryMetricsFetchEligibility(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	eligible, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, eligible)
}

func Test_CheckWaitEventMetricsFetchEligibility(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_wait_sampling'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	eligible, err := CheckWaitEventMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, eligible)
}

func Test_CheckBlockingSessionMetricsFetchEligibility(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	eligible, err := CheckBlockingSessionMetricsFetchEligibility(conn, 0)
	assert.NoError(t, err)
	assert.True(t, eligible)
}

func Test_CheckIndividualQueryMetricsFetchEligibility(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	eligible, err := CheckIndividualQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, eligible)
}
