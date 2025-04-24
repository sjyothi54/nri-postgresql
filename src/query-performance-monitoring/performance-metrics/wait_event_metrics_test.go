package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	enabledExtensions := map[string]bool{
		"pg_wait_sampling":   true,
		"pg_stat_statements": true,
	}
	version := uint64(14)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	expectedQuery := queries.WaitEvents
	query := fmt.Sprintf(expectedQuery, databaseName, args.QueryMonitoringCountThreshold)
	rowData := []driver.Value{
		"Locks:Lock", "Locks", 500.0, "2023-01-01T00:00:00Z", "queryid2", "SELECT 2", "testdb",
	}
	expectedRows := [][]driver.Value{
		rowData, rowData,
	}
	mockRows := sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(rowData...).AddRow(rowData...)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(mockRows)
	waitEventMetricsList, err := getWaitEventMetrics(conn, cp, enabledExtensions)
	compareMockRowsWithWaitMetrics(t, expectedRows, waitEventMetricsList)
	assert.NoError(t, err)
	assert.Len(t, waitEventMetricsList, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func compareMockRowsWithWaitMetrics(t *testing.T, expectedRows [][]driver.Value, waitEventMetricsList []any) {
	assert.Equal(t, 2, len(waitEventMetricsList))
	for index := range waitEventMetricsList {
		waitEvents := waitEventMetricsList[index].(datamodels.WaitEventMetrics)
		assert.Equal(t, expectedRows[index][0], *waitEvents.WaitEventName)
		assert.Equal(t, expectedRows[index][1], *waitEvents.WaitCategory)
		assert.Equal(t, expectedRows[index][2], *waitEvents.TotalWaitTimeMs)
		assert.Equal(t, expectedRows[index][3], *waitEvents.CollectionTimestamp)
		assert.Equal(t, expectedRows[index][4], *waitEvents.QueryID)
		assert.Equal(t, expectedRows[index][5], *waitEvents.QueryText)
		assert.Equal(t, expectedRows[index][6], *waitEvents.DatabaseName)
	}
}
func TestGetWaitEventMetricsFromPgStatActivity(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"

	cp := common_parameters.SetCommonParameters(args, uint64(14), databaseName)

	enabledExtensions := map[string]bool{
		"pg_wait_sampling":   false,
		"pg_stat_statements": true,
	}
	query := fmt.Sprintf(queries.WaitEventsFromPgStatActivity, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 500.0, "2023-01-01T00:00:00Z", "queryid2", "SELECT 2", "testdb",
	))
	waitEventsList, err := getWaitEventMetrics(conn, cp, enabledExtensions)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 1)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetWaitEventEmptyMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	enabledExtensions := map[string]bool{
		"pg_wait_sampling":   true,
		"pg_stat_statements": true,
	}
	version := uint64(14)
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	expectedQuery := queries.WaitEvents
	query := fmt.Sprintf(expectedQuery, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}))
	waitEventsList, err := getWaitEventMetrics(conn, cp, enabledExtensions)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}
