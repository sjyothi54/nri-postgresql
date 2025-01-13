package performancemetrics

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PopulateExecutionPlanMetrics_NoIndividualQueriesFound(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}
	results := []datamodels.IndividualQueryMetrics{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
	// No assertion needed as the function should just log and return
}

func Test_PopulateExecutionPlanMetrics_WithIndividualQueries(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "test"}}]`))

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:       strPtr("1"),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
}

func Test_PopulateExecutionPlanMetrics_ExecutionPlanNotFound(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}))

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:       strPtr("1"),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
	// No assertion needed as the function should just log and return
}

func Test_PopulateExecutionPlanMetrics_ErrorScanningRow(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(nil).RowError(0, sqlmock.ErrCancelled))

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:       strPtr("1"),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
	// No assertion needed as the function should just log and return
}

func Test_PopulateExecutionPlanMetrics_FailedToUnmarshal(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`invalid json`))

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:       strPtr("1"),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
	// No assertion needed as the function should just log and return
}

func Test_PopulateExecutionPlanMetrics_Success(t *testing.T) {
	conn, mock := performanceDbConnection.CreateMockSQL(t)
	defer conn.Close()

	mock.ExpectQuery("EXPLAIN \\(FORMAT JSON\\) SELECT \\* FROM test").WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "test"}}]`))

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:       strPtr("1"),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test"),
			RealQueryText: strPtr("SELECT * FROM test"),
		},
	}

	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)
	// No assertion needed as the function should just log and return
}

func Test_fetchNestedExecutionPlanDetails_Success(t *testing.T) {
	individualQuery := datamodels.IndividualQueryMetrics{
		QueryID:       strPtr("1"),
		DatabaseName:  strPtr("testdb"),
		QueryText:     strPtr("SELECT * FROM test"),
		RealQueryText: strPtr("SELECT * FROM test"),
		PlanID:        strPtr("123"),
	}

	execPlan := map[string]interface{}{
		"Node Type": "Seq Scan",
		"Plans": []interface{}{
			map[string]interface{}{
				"Node Type": "Index Scan",
			},
		},
	}

	var executionPlanMetricsList []interface{}
	level := 0
	fetchNestedExecutionPlanDetails(individualQuery, &level, execPlan, &executionPlanMetricsList)

	assert.Len(t, executionPlanMetricsList, 2)
	assert.Equal(t, "Seq Scan", executionPlanMetricsList[0].(datamodels.QueryExecutionPlanMetrics).NodeType)
	assert.Equal(t, "Index Scan", executionPlanMetricsList[1].(datamodels.QueryExecutionPlanMetrics).NodeType)
	assert.Equal(t, "1", executionPlanMetricsList[0].(datamodels.QueryExecutionPlanMetrics).QueryID)
	assert.Equal(t, "123", executionPlanMetricsList[0].(datamodels.QueryExecutionPlanMetrics).PlanID)
	assert.Equal(t, 0, executionPlanMetricsList[0].(datamodels.QueryExecutionPlanMetrics).Level)
	assert.Equal(t, 1, executionPlanMetricsList[1].(datamodels.QueryExecutionPlanMetrics).Level)
}
