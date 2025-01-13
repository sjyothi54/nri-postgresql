package performancemetrics

import (
	"errors"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_GetBlockingMetrics_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `SELECT * FROM pg_stat_activity WHERE state = 'active' AND wait_event IS NOT NULL`
	mock.ExpectQuery(query).WillReturnError(errors.New("query error"))

	args := args.ArgumentList{}
	metrics, err := GetBlockingMetrics(conn, args, "someString", 12345)
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func Test_GetBlockingMetrics_CloseError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `SELECT
          blocked_activity.pid AS blocked_pid,
          blocked_statements.query AS blocked_query,
          blocked_statements.queryid AS blocked_query_id,
          blocked_activity.query_start AS blocked_query_start,
          blocked_activity.datname AS database_name,
          blocking_activity.pid AS blocking_pid,
          blocking_statements.query AS blocking_query,
          blocking_statements.queryid AS blocking_query_id,
          blocking_activity.query_start AS blocking_query_start
      FROM pg_stat_activity AS blocked_activity
      JOIN pg_stat_statements as blocked_statements on blocked_activity.query_id = blocked_statements.queryid
      JOIN pg_locks blocked_locks ON blocked_activity.pid = blocked_locks.pid
      JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
          AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database
          AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation
          AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page
          AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple
          AND blocked_locks.transactionid IS NOT DISTINCT FROM blocking_locks.transactionid
          AND blocked_locks.classid IS NOT DISTINCT FROM blocking_locks.classid
          AND blocked_locks.objid IS NOT DISTINCT FROM blocking_locks.objid
          AND blocked_locks.objsubid IS NOT DISTINCT FROM blocking_locks.objsubid
          AND blocked_locks.pid <> blocking_locks.pid
      JOIN pg_stat_activity AS blocking_activity ON blocking_locks.pid = blocking_activity.pid
      JOIN pg_stat_statements as blocking_statements on blocking_activity.query_id = blocking_statements.queryid
      WHERE NOT blocked_locks.granted
          AND blocked_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%'
          AND blocking_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%'
      LIMIT 10;`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name", "blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start"}).
		AddRow(1, "SELECT * FROM test", 12345, "2023-10-10 10:10:10", "testdb", 2, "UPDATE test SET value = 1", 67890, "2023-10-10 10:10:20"))

	mock.ExpectClose().WillReturnError(errors.New("close error"))

	args := args.ArgumentList{}
	metrics, err := GetBlockingMetrics(conn, args, "someString", 12345)
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func Test_PopulateBlockingMetrics_NoExtension(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateBlockingMetrics(conn, pgIntegration, args, "someString", 12345)
}

func Test_PopulateBlockingMetrics_IngestMetrics(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_extension WHERE extname = 'pg_stat_statements'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	query := `SELECT
          blocked_activity.pid AS blocked_pid,
          blocked_statements.query AS blocked_query,
          blocked_statements.queryid AS blocked_query_id,
          blocked_activity.query_start AS blocked_query_start,
          blocked_activity.datname AS database_name,
          blocking_activity.pid AS blocking_pid,
          blocking_statements.query AS blocking_query,
          blocking_statements.queryid AS blocking_query_id,
          blocking_activity.query_start AS blocking_query_start
      FROM pg_stat_activity AS blocked_activity
      JOIN pg_stat_statements as blocked_statements on blocked_activity.query_id = blocked_statements.queryid
      JOIN pg_locks blocked_locks ON blocked_activity.pid = blocked_locks.pid
      JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
          AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database
          AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation
          AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page
          AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple
          AND blocked_locks.transactionid IS NOT DISTINCT FROM blocking_locks.transactionid
          AND blocked_locks.classid IS NOT DISTINCT FROM blocking_locks.classid
          AND blocked_locks.objid IS NOT DISTINCT FROM blocking_locks.objid
          AND blocked_locks.objsubid IS NOT DISTINCT FROM blocking_locks.objsubid
          AND blocked_locks.pid <> blocking_locks.pid
      JOIN pg_stat_activity AS blocking_activity ON blocking_locks.pid = blocking_activity.pid
      JOIN pg_stat_statements as blocking_statements on blocking_activity.query_id = blocking_statements.queryid
      WHERE NOT blocked_locks.granted
          AND blocked_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%'
          AND blocking_statements.query NOT LIKE 'EXPLAIN (FORMAT JSON) %%'
      LIMIT 10;`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name", "blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start"}).
		AddRow(1, "SELECT * FROM test", 12345, "2023-10-10 10:10:10", "testdb", 2, "UPDATE test SET value = 1", 67890, "2023-10-10 10:10:20"))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateBlockingMetrics(conn, pgIntegration, args, "someString", 12345)

}
