package performance_metrics

import (
	"errors"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_GetSlowRunningMetrics_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema() AS schema_name, pss.calls AS execution_count, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN (FORMAT JSON) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`
	mock.ExpectQuery(query).WillReturnError(errors.New("query error"))

	args := args.ArgumentList{}
	metrics, metricsInterface, err := GetSlowRunningMetrics(conn, args)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Nil(t, metricsInterface)
}
func Test_PopulateSlowRunningMetrics_NoExtension(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	metrics := PopulateSlowRunningMetrics(conn, pgIntegration, args)
	assert.Nil(t, metrics)
}
func Test_PopulateSlowRunningMetrics_NoSlowQueries(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema\(\) AS schema_name, pss.calls AS execution_count, ROUND\(\(pss.total_exec_time / pss.calls\)::numeric, 3\) AS avg_elapsed_time_ms, ROUND\(\(pss.total_exec_time / pss.calls\)::numeric, 3\) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char\(NOW\(\) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"\) AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN \(FORMAT JSON\) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`).
		WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name", "schema_name", "execution_count", "avg_elapsed_time_ms", "avg_cpu_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp"}))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowQueryMetricsList := PopulateSlowRunningMetrics(conn, pgIntegration, args)
	assert.Nil(t, slowQueryMetricsList)
}
func Test_PopulateSlowRunningMetrics_ErrorFetching(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema\(\) AS schema_name, pss.calls AS execution_count, ROUND\(\(pss.total_exec_time / pss.calls\)::numeric, 3\) AS avg_elapsed_time_ms, ROUND\(\(pss.total_exec_time / pss.calls\)::numeric, 3\) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char\(NOW\(\) AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"\) AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN \(FORMAT JSON\) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`).
		WillReturnError(errors.New("query error"))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowQueryMetricsList := PopulateSlowRunningMetrics(conn, pgIntegration, args)
	assert.Nil(t, slowQueryMetricsList)
}
func Test_GetSlowRunningMetrics_StructScanError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema() AS schema_name, pss.calls AS execution_count, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN (FORMAT JSON) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`
	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name", "schema_name", "execution_count", "avg_elapsed_time_ms", "avg_cpu_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp"}).
		AddRow("invalid_id", "SELECT * FROM test", "testdb", "public", 10, 1500, 1500, 0, 0, "SELECT", "2023-10-10T10:10:10Z"))

	args := args.ArgumentList{}
	metrics, metricsInterface, err := GetSlowRunningMetrics(conn, args)
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Nil(t, metricsInterface)
}
func Test_GetSlowRunningMetrics_Success(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema() AS schema_name, pss.calls AS execution_count, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN (FORMAT JSON) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name", "schema_name", "execution_count", "avg_elapsed_time_ms", "avg_cpu_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp"}).
		AddRow(1, "SELECT * FROM test", "testdb", "public", 10, 1500, 1500, 0, 0, "SELECT", "2023-10-10T10:10:10Z"))

	args := args.ArgumentList{}
	metrics, metricsInterface, err := GetSlowRunningMetrics(conn, args)
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metricsInterface)
	assert.Equal(t, 1, len(metrics))
	assert.Equal(t, 1, len(metricsInterface))
}
func Test_PopulateSlowRunningMetrics_WithSlowQueries(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	query := `SELECT pss.queryid AS query_id, pss.query AS query_text, pd.datname AS database_name, current_schema() AS schema_name, pss.calls AS execution_count, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms, ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms, pss.shared_blks_read / pss.calls AS avg_disk_reads, pss.shared_blks_written / pss.calls AS avg_disk_writes, CASE WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT' WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT' WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE' WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE' ELSE 'OTHER' END AS statement_type, to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp FROM pg_stat_statements pss JOIN pg_database pd ON pss.dbid = pd.oid WHERE pss.query NOT LIKE 'EXPLAIN (FORMAT JSON) %' ORDER BY avg_elapsed_time_ms DESC LIMIT 20`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name", "schema_name", "execution_count", "avg_elapsed_time_ms", "avg_cpu_time_ms", "avg_disk_reads", "avg_disk_writes", "statement_type", "collection_timestamp"}).
		AddRow(1, "SELECT * FROM test", "testdb", "public", 10, 1500, 1500, 0, 0, "SELECT", "2023-10-10T10:10:10Z"))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowQueryMetricsList := PopulateSlowRunningMetrics(conn, pgIntegration, args)
	assert.NotNil(t, slowQueryMetricsList)
	assert.Equal(t, 1, len(slowQueryMetricsList))
}
