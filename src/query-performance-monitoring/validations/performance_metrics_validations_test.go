package validations

import (
	"errors"
	"testing"

	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_CheckPgWaitSamplingExtensionEnabled_Enabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_wait_sampling'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgWaitSamplingExtensionEnabled_Disabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_wait_sampling'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgWaitSamplingExtensionEnabled_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_wait_sampling'").
		WillReturnError(errors.New("query error"))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.Error(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatStatementsExtensionEnabled_Enabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatStatementsExtensionEnabled_Disabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatStatementsExtensionEnabled_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").
		WillReturnError(errors.New("query error"))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.Error(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatMonitorExtensionEnabled_Enabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.True(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatMonitorExtensionEnabled_Disabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.NoError(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CheckPgStatMonitorExtensionEnabled_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnError(errors.New("query error"))

	enabled, err := CheckSlowQueryMetricsFetchEligibility(conn)
	assert.Error(t, err)
	assert.False(t, enabled)
	assert.NoError(t, mock.ExpectationsWereMet())
}
