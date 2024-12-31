package performancemetrics

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PopulateExecutionPlanMetrics_NoIndividualQueriesFound(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(nil, pgIntegration, args)
	// No assertion needed, just ensuring no panic occurs
}

func Test_PopulateExecutionPlanMetrics_Success(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "test"}}]`))

	individualQueries := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
			PlanID:       strPtr("plan1"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	// No assertion needed, just ensuring no panic occurs
}

func Test_ProcessExecutionPlanOfQueries_NoRows(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}))

	individualQueries := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
			PlanID:       strPtr("plan1"),
		},
	}

	var executionPlanMetricsList []interface{}
	processExecutionPlanOfQueries(individualQueries, conn, &executionPlanMetricsList)
	assert.Empty(t, executionPlanMetricsList)
}

func Test_ProcessExecutionPlanOfQueries_ScanError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(nil))

	individualQueries := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
			PlanID:       strPtr("plan1"),
		},
	}

	var executionPlanMetricsList []interface{}
	processExecutionPlanOfQueries(individualQueries, conn, &executionPlanMetricsList)
	assert.Empty(t, executionPlanMetricsList)
}

func Test_ProcessExecutionPlanOfQueries_UnmarshalError(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`invalid json`))

	individualQueries := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
			PlanID:       strPtr("plan1"),
		},
	}

	var executionPlanMetricsList []interface{}
	processExecutionPlanOfQueries(individualQueries, conn, &executionPlanMetricsList)
	assert.Empty(t, executionPlanMetricsList)
}

func Test_ProcessExecutionPlanOfQueries_Success(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "test"}}]`))

	individualQueries := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      int64Ptr(1),
			DatabaseName: strPtr("testdb"),
			QueryText:    strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
			PlanID:       strPtr("plan1"),
		},
	}

	var executionPlanMetricsList []interface{}
	processExecutionPlanOfQueries(individualQueries, conn, &executionPlanMetricsList)
	assert.NotEmpty(t, executionPlanMetricsList)
	assert.Equal(t, "Seq Scan", executionPlanMetricsList[0].(datamodels.QueryExecutionPlanMetrics).NodeType)
}