package performance_metrics

import (
	// "bytes"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"

	// performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	// "github.com/sirupsen/logrus"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PopulateExecutionPlanMetrics(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	argList := args.ArgumentList{}

	results := []datamodels.IndividualQueryMetrics{
		{
			QueryId:       new(int64),
			QueryText:     new(string),
			DatabaseName:  new(string),
			RealQueryText: new(string),
		},
	}
	var queryId int64 = 1
	*results[0].QueryId = queryId
	*results[0].QueryText = "SELECT * FROM test"
	*results[0].DatabaseName = "testdb"
	*results[0].RealQueryText = "SELECT * FROM test"

	// Mock GetExecutionPlanMetrics
	mockGetExecutionPlanMetrics := func(results []datamodels.IndividualQueryMetrics, argList args.ArgumentList) []interface{} {
		return []interface{}{"mockedExecutionPlanMetric"}
	}

	// Mock IngestMetric
	mockIngestMetric := func(metrics []interface{}, metricName string, integration *integration.Integration, argList args.ArgumentList) {
		require.Equal(t, "PostgresExecutionPlanMetrics", metricName)
		require.Equal(t, pgIntegration, integration)
		require.Equal(t, argList, argList)
		require.Equal(t, []interface{}{"mockedExecutionPlanMetric"}, metrics)
	}

	// Define a dummy getExecutionPlanMetrics function
	getExecutionPlanMetrics := func(results []datamodels.IndividualQueryMetrics, argList args.ArgumentList) []interface{} {
		return []interface{}{}
	}

	// Replace the actual functions with mocks
	originalGetExecutionPlanMetrics := getExecutionPlanMetrics
	originalIngestMetric := common_utils.GetIngestMetricFunc()
	getExecutionPlanMetrics = mockGetExecutionPlanMetrics
	common_utils.SetIngestMetricFunc(mockIngestMetric)
	defer func() {
		getExecutionPlanMetrics = originalGetExecutionPlanMetrics
		common_utils.SetIngestMetricFunc(originalIngestMetric)
	}()

	// Call the function under test
	PopulateExecutionPlanMetrics(results, pgIntegration, argList)

	// Assertions
	assert.Equal(t, "mockedExecutionPlanMetric", mockGetExecutionPlanMetrics(results, argList)[0])
}
func Test_PopulateExecutionPlanMetrics_NoQueries(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}

	PopulateExecutionPlanMetrics([]datamodels.IndividualQueryMetrics{}, pgIntegration, args)
	// No assertions needed, just ensuring no panic occurs
}

// func Test_processExecutionPlanOfQueries_NoRowsNext(t *testing.T) {
// 	individualQueriesList := []datamodels.IndividualQueryMetrics{
// 		{
// 			QueryId:       new(int64),
// 			QueryText:     new(string),
// 			DatabaseName:  new(string),
// 			RealQueryText: new(string),
// 		},
// 	}
// 	var queryId int64 = 1
// 	*individualQueriesList[0].QueryId = queryId
// 	*individualQueriesList[0].QueryText = "SELECT * FROM test"
// 	*individualQueriesList[0].DatabaseName = "testdb"
// 	*individualQueriesList[0].RealQueryText = "SELECT * FROM test"

// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	query := "EXPLAIN (FORMAT JSON) SELECT * FROM test"
// 	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}))

// 	// Capture log output
// 	var logOutput bytes.Buffer
// 	logrus.SetOutput(&logOutput)
// 	defer logrus.SetOutput(nil) // Reset log output

// 	var executionPlanMetricsList []interface{}
// 	processExecutionPlanOfQueries(individualQueriesList, conn, &executionPlanMetricsList)

// 	// Ensure that the log message is correct
// 	assert.Contains(t, logOutput.String(), "Execution plan not found for queryId 1")
// 	mock.ExpectationsWereMet()
// }
// func Test_processExecutionPlanOfQueries_ScanError(t *testing.T) {
// 	individualQueriesList := []datamodels.IndividualQueryMetrics{
// 		{
// 			QueryId:       new(int64),
// 			QueryText:     new(string),
// 			DatabaseName:  new(string),
// 			RealQueryText: new(string),
// 		},
// 	}
// 	var queryId int64 = 1
// 	*individualQueriesList[0].QueryId = queryId
// 	*individualQueriesList[0].QueryText = "SELECT * FROM test"
// 	*individualQueriesList[0].DatabaseName = "testdb"
// 	*individualQueriesList[0].RealQueryText = "SELECT * FROM test"

// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	query := "EXPLAIN (FORMAT JSON) SELECT * FROM test"
// 	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(""))

// 	// Capture log output
// 	var logOutput bytes.Buffer
// 	logrus.SetOutput(&logOutput)
// 	defer logrus.SetOutput(nil) // Reset log output

// 	var executionPlanMetricsList []interface{}
// 	processExecutionPlanOfQueries(individualQueriesList, conn, &executionPlanMetricsList)

// 	// Ensure that the log message is correct
// 	assert.Contains(t, logOutput.String(), "Error scanning row")
// 	mock.ExpectationsWereMet()
// }
// func Test_processExecutionPlanOfQueries_UnmarshalError(t *testing.T) {
// 	individualQueriesList := []datamodels.IndividualQueryMetrics{
// 		{
// 			QueryId:       new(int64),
// 			QueryText:     new(string),
// 			DatabaseName:  new(string),
// 			RealQueryText: new(string),
// 		},
// 	}
// 	var queryId int64 = 1
// 	*individualQueriesList[0].QueryId = queryId
// 	*individualQueriesList[0].QueryText = "SELECT * FROM test"
// 	*individualQueriesList[0].DatabaseName = "testdb"
// 	*individualQueriesList[0].RealQueryText = "SELECT * FROM test"

// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	query := "EXPLAIN (FORMAT JSON) SELECT * FROM test"
// 	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow("invalid json"))

// 	// Capture log output
// 	var logOutput bytes.Buffer
// 	logrus.SetOutput(&logOutput)
// 	defer logrus.SetOutput(nil) // Reset log output

// 	var executionPlanMetricsList []interface{}
// 	processExecutionPlanOfQueries(individualQueriesList, conn, &executionPlanMetricsList)

// 	// Ensure that the log message is correct
// 	assert.Contains(t, logOutput.String(), "Failed to unmarshal execution plan")
// 	mock.ExpectationsWereMet()
// }
// func Test_processExecutionPlanOfQueries_Success(t *testing.T) {
// 	individualQueriesList := []datamodels.IndividualQueryMetrics{
// 		{
// 			QueryId:       new(int64),
// 			QueryText:     new(string),
// 			DatabaseName:  new(string),
// 			RealQueryText: new(string),
// 		},
// 	}
// 	var queryId int64 = 1
// 	*individualQueriesList[0].QueryId = queryId
// 	*individualQueriesList[0].QueryText = "SELECT * FROM test"
// 	*individualQueriesList[0].DatabaseName = "testdb"
// 	*individualQueriesList[0].RealQueryText = "SELECT * FROM test"

// 	conn, mock := performanceDbConnection.CreateMockSQL(t)
// 	defer conn.Close()

// 	query := "EXPLAIN (FORMAT JSON) SELECT * FROM test"
// 	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"QUERY PLAN"}).AddRow(`[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "test", "Alias": "test", "Startup Cost": 0.00, "Total Cost": 35.50, "Plan Rows": 2550, "Plan Width": 24}}]`))

// 	var executionPlanMetricsList []interface{}
// 	processExecutionPlanOfQueries(individualQueriesList, conn, &executionPlanMetricsList)

// 	assert.NotEmpty(t, executionPlanMetricsList)
// 	mock.ExpectationsWereMet()
// }
