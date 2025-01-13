package performancemetrics

import (
	"errors"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_GetWaitEventMetrics_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := queries.WaitEvents
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errors.New("query error"))

	argList := args.ArgumentList{}
	metrics, err := GetWaitEventMetrics(conn, argList, "someString")
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func Test_GetWaitEventMetrics_ScanError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	query := `WITH wait_history AS (
        SELECT
            wh.pid,
            wh.event_type,
            wh.event,
            wh.ts,
            pg_database.datname AS database_name,
            LEAD(wh.ts) OVER (PARTITION BY wh.pid ORDER BY wh.ts) - wh.ts AS duration,
            sa.query AS query_text,
            sa.queryid AS query_id
        FROM
            pg_wait_sampling_history wh
        LEFT JOIN
            pg_stat_statements sa ON wh.queryid = sa.queryid
        LEFT JOIN
            pg_database ON pg_database.oid = sa.dbid
    )
    SELECT
        event_type || ':' || event AS wait_event_name,
        CASE
            WHEN event_type IN ('LWLock', 'Lock') THEN 'Locks'
            WHEN event_type = 'IO' THEN 'Disk IO'
            WHEN event_type = 'CPU' THEN 'CPU'
            ELSE 'Other'
        END AS wait_category,
        EXTRACT(EPOCH FROM SUM(duration)) * 1000 AS total_wait_time_ms,  -- Convert duration to milliseconds
        COUNT(DISTINCT pid) AS waiting_tasks_count,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp,
        query_id,
        query_text,
        database_name
    FROM wait_history
    WHERE query_text NOT LIKE 'EXPLAIN (FORMAT JSON) %' AND query_id IS NOT NULL AND event_type IS NOT NULL
    GROUP BY event_type, event, query_id, query_text, database_name
    ORDER BY total_wait_time_ms DESC`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"wait_event_name", "wait_category", "total_wait_time_ms", "waiting_tasks_count", "collection_timestamp", "query_id", "query_text", "database_name"}).
		AddRow("Lock:relation", "Locks", 1000, 1, "2023-10-10T10:10:10Z", 12345, "SELECT * FROM test", "testdb").
		AddRow("Lock:relation", "Locks", "invalid", 1, "2023-10-10T10:10:10Z", 12345, "SELECT * FROM test", "testdb"))

	argList := args.ArgumentList{}
	metrics, err := GetWaitEventMetrics(conn, argList, "someString")
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func Test_PopulateWaitEventMetrics_ExtensionError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT .* FROM pg_extension WHERE extname = 'pg_wait_sampling'").WillReturnError(errors.New("query error"))

	pgIntegration, _ := integration.New("test", "1.0.0")
	argList := args.ArgumentList{}
	PopulateWaitEventMetrics(conn, pgIntegration, argList, "someString")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_PopulateWaitEventMetrics_ExtensionNotEnabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT .* FROM pg_extension WHERE extname = 'pg_wait_sampling'").WillReturnRows(sqlmock.NewRows([]string{"extname"}))

	pgIntegration, _ := integration.New("test", "1.0.0")
	argList := args.ArgumentList{}
	PopulateWaitEventMetrics(conn, pgIntegration, argList, "someString")

	assert.NoError(t, mock.ExpectationsWereMet())
}
