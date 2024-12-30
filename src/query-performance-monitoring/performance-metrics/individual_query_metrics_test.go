package performance_metrics

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PopulateIndividualQueryMetrics_ExtensionNotEnabled(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	result := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.Nil(t, result)
}

func Test_PopulateIndividualQueryMetrics_NoIndividualQueriesFound(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT query, queryid, datname, planid, ROUND\\(\\(total_exec_time / calls\\)::numeric, 3\\) AS avg_elapsed_time_ms, ROUND\\(\\(\\(cpu_user_time \\+ cpu_sys_time\\) / NULLIF\\(calls, 0\\)\\)::numeric, 3\\) AS avg_cpu_time_ms FROM pg_stat_monitor WHERE queryid IN \\(1\\) AND avg_elapsed_time_ms > 0 AND bucket_start_time >= NOW\\(\\) - INTERVAL '15 seconds' GROUP BY query, queryid, datname, planid, total_exec_time, cpu_user_time, cpu_sys_time, calls").WillReturnRows(sqlmock.NewRows([]string{"query", "queryid", "datname", "planid", "avg_elapsed_time_ms", "avg_cpu_time_ms"}))

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	result := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.Nil(t, result)
}

func Test_PopulateIndividualQueryMetrics_Success(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT query, queryid, datname, planid, ROUND\\(\\(total_exec_time / calls\\)::numeric, 3\\) AS avg_elapsed_time_ms, ROUND\\(\\(\\(cpu_user_time \\+ cpu_sys_time\\) / NULLIF\\(calls, 0\\)\\)::numeric, 3\\) AS avg_cpu_time_ms FROM pg_stat_monitor WHERE queryid IN \\(1\\) AND avg_elapsed_time_ms > 0 AND bucket_start_time >= NOW\\(\\) - INTERVAL '15 seconds' GROUP BY query, queryid, datname, planid, total_exec_time, cpu_user_time, cpu_sys_time, calls").WillReturnRows(sqlmock.NewRows([]string{"query", "queryid", "datname", "planid", "avg_cpu_time_ms"}).
		AddRow("SELECT * FROM test", 1, "testdb", 1, 5.0))

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	result := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.NotNil(t, result)
	assert.Equal(t, "SELECT * FROM test", *result[0].RealQueryText)
}
