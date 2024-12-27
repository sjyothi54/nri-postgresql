package performance_metrics

import (
	"errors"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PopulateIndividualQueryMetrics_NoExtension(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{}
	metrics := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.Nil(t, metrics)
}

func Test_PopulateIndividualQueryMetrics_ErrorFetchingExtension(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnError(errors.New("query error"))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{}
	metrics := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.Nil(t, metrics)
}

func Test_PopulateIndividualQueryMetrics_NoQueries(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(queries.IndividualQuerySearch).
		WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name"}))

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{}
	metrics := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
	assert.Nil(t, metrics)
}

// func Test_PopulateIndividualQueryMetrics_Success(t *testing.T) {
// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").
// 		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

// 	mock.ExpectQuery(queries.IndividualQuerySearch).
// 		WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name"}).
// 			AddRow(1, "SELECT * FROM test", "testdb"))

// 	pgIntegration, _ := integration.New("test", "1.0.0")
// 	args := args.ArgumentList{}

// 	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
// 		{
// 			QueryID:      new(int64),
// 			QueryText:    new(string),
// 			DatabaseName: new(string),
// 		},
// 	}

// 	metrics := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args)
// 	assert.NotNil(t, metrics)
// 	assert.Equal(t, 1, len(metrics))
// }

func Test_GetIndividualQueryMetrics_Error(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery(queries.IndividualQuerySearch).
		WillReturnError(errors.New("query error"))

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      new(int64),
			QueryText:    new(string),
			DatabaseName: new(string),
		},
	}
	*slowRunningQueries[0].QueryID = 1
	*slowRunningQueries[0].QueryText = "SELECT * FROM test"
	*slowRunningQueries[0].DatabaseName = "testdb"

	args := args.ArgumentList{}
	metrics, metricsForExecPlan := GetIndividualQueryMetrics(conn, args, slowRunningQueries)
	assert.Nil(t, metrics)
	assert.Nil(t, metricsForExecPlan)
}

// func Test_GetIndividualQueryMetrics_Success(t *testing.T) {
// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	mock.ExpectQuery(queries.IndividualQuerySearch).
// 		WillReturnRows(sqlmock.NewRows([]string{"query_id", "query_text", "database_name"}).
// 			AddRow(1, "SELECT * FROM test", "testdb"))

// 	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
// 		{
// 			QueryID:      new(int64),
// 			QueryText:    new(string),
// 			DatabaseName: new(string),
// 		},
// 	}
// 	*slowRunningQueries[0].QueryID = 1
// 	*slowRunningQueries[0].QueryText = "SELECT * FROM test"
// 	*slowRunningQueries[0].DatabaseName = "testdb"

// 	metrics, metricsForExecPlan := GetIndividualQueryMetrics(conn, slowRunningQueries)
// 	assert.NotNil(t, metrics)
// 	assert.NotNil(t, metricsForExecPlan)
// 	assert.Equal(t, 1, len(metrics))
// 	assert.Equal(t, 1, len(metricsForExecPlan))
// }
