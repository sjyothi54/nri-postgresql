package query_results

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/validations"
)

// FetchAndLogExecutionPlan fetches the execution plan for a given query text and logs the result
func FetchAndLogExecutionPlan(conn *connection.PGSQLConnection, queryText string) (string, error) {
	var executionPlan string
	query := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", queryText)
	err := conn.QueryRowx(query).Scan(&executionPlan)
	if err != nil {
		log.Error("Error fetching execution plan for query: %v", err)
		return "", err
	}
	log.Info("Execution Plan for Query: %s", executionPlan)
	return executionPlan, nil
}

func GetQueryExecutionPlanMetrics(conn *connection.PGSQLConnection) ([]datamodels.QueryExecutionPlan, error) {
	var slowQueries []datamodels.QueryExecutionPlan
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slowQuery datamodels.QueryExecutionPlan
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
	}

	return slowQueries, nil
}

// PopulateSlowQueryMetrics fetches slow-running metrics and returns the list of query texts
func PopulateSlowQueryMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, args interface{}) ([]string, error) {
	slowQueries, err := GetQueryExecutionPlanMetrics(conn)
	if err != nil {
		return nil, err
	}

	var queryTextList []string
	for _, query := range slowQueries {
		queryTextList = append(queryTextList, *query.QueryText)
	}

	return queryTextList, nil
}

// PopulateIndividualQueryDetails fetches individual query details based on the query text list
func PopulateIndividualQueryDetails(conn *connection.PGSQLConnection, queryTextList []string, instanceEntity *integration.Entity, args interface{}) ([]string, error) {
	var individualQueryDetails []string
	for _, queryText := range queryTextList {
		queryDetails, err := FetchAndLogExecutionPlan(conn, queryText)
		if err != nil {
			return nil, err
		}
		individualQueryDetails = append(individualQueryDetails, queryDetails)
	}
	return individualQueryDetails, nil
}

// PopulateExecutionPlans fetches execution plans based on the individual query details
func PopulateExecutionPlans(conn *connection.PGSQLConnection, individualQueryDetails []string, instanceEntity *integration.Entity, args interface{}) ([]string, error) {
	var executionPlanMetrics []string
	for _, queryDetail := range individualQueryDetails {
		executionPlan, err := FetchAndLogExecutionPlan(conn, queryDetail)
		if err != nil {
			return nil, err
		}
		executionPlanMetrics = append(executionPlanMetrics, executionPlan)
	}
	return executionPlanMetrics, nil
}

// Main function to call the above functions and log the results
func MainFunction(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, args interface{}) {
	queryTextList, err := PopulateSlowQueryMetrics(instanceEntity, conn, args)
	if err != nil {
		log.Error("Error populating slow query metrics: %v", err)
		return
	}

	individualQueryDetails, err := PopulateIndividualQueryDetails(conn, queryTextList, instanceEntity, args)
	if err != nil {
		log.Error("Error populating individual query details: %v", err)
		return
	}
	fmt.Println("Query Plan details collected successfully.", individualQueryDetails)

	executionPlanMetrics, err := PopulateExecutionPlans(conn, individualQueryDetails, instanceEntity, args)
	if err != nil {
		log.Error("Error populating execution plan details: %v", err)
		return
	}
	fmt.Println("Execution plan details collected successfully.", executionPlanMetrics)

	// Set up the results JSON data in the event PostgresSQLQueryPlanGo
	for _, executionPlan := range executionPlanMetrics {
		metricSet := instanceEntity.NewMetricSet("PostgresSQLQueryPlanGo")
		var planData []map[string]interface{}
		if err := json.Unmarshal([]byte(executionPlan), &planData); err != nil {
			log.Error("Error unmarshalling execution plan JSON: %v", err)
			continue
		}
		for _, plan := range planData {
			for key, value := range plan {
				setQueryExecutionMetrics(metricSet, key, value, "attribute")
			}
		}
	}
}

// PopulateQueryExecutionMetrics fetches slow-running metrics and populates them into the metric set
func PopulateQueryExecutionMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' enabled.")
		executionPlan, err := GetQueryExecutionPlanMetrics(conn)
		if err != nil {
			log.Error("Error fetching slow-running queries: %v", err)
			return
		}

		if len(executionPlan) == 0 {
			log.Info("No slow-running queries found.")
			return
		}
		log.Info("Populate-slow running: %+v", executionPlan)

		for _, model := range executionPlan {
			metricSet := instanceEntity.NewMetricSet("PostgresSQLQueryPlanGo")

			modelValue := reflect.ValueOf(model)
			modelType := reflect.TypeOf(model)

			for i := 0; i < modelValue.NumField(); i++ {
				field := modelValue.Field(i)
				fieldType := modelType.Field(i)
				metricName := fieldType.Tag.Get("metric_name")
				sourceType := fieldType.Tag.Get("source_type")

				if field.Kind() == reflect.Ptr && !field.IsNil() {
					setQueryExecutionMetrics(metricSet, metricName, field.Elem().Interface(), sourceType)
				} else if field.Kind() != reflect.Ptr {
					setQueryExecutionMetrics(metricSet, metricName, field.Interface(), sourceType)
				}
			}

			log.Info("Metrics set for slow query: %d in database: %s", *model.QueryID, *model.DatabaseName)
		}
	} else {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}

}

func setQueryExecutionMetrics(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		metricSet.SetMetric(name, value, metric.GAUGE)
	case `attribute`:
		metricSet.SetMetric(name, value, metric.ATTRIBUTE)
	default:
		metricSet.SetMetric(name, value, metric.GAUGE)
	}
}
