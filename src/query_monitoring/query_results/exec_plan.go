package query_results

import (
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
)

func ExecutionPlan(conn *connection.PGSQLConnection) ([]datamodels.ExecutionPlan, error) {
	var query = queries.ExecutionPlanQuery
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var executionPlans []datamodels.ExecutionPlan
	for rows.Next() {
		var executionPlan datamodels.ExecutionPlan
		if err := rows.Scan(&executionPlan.Query, &executionPlan.QueryID); err != nil {
			log.Error("Error scanning row: %v", err)
			return nil, err
		}
		log.Info("query: %s queryid: %s", executionPlan.Query, executionPlan.QueryID)
		executionPlans = append(executionPlans, executionPlan)
	}
	return executionPlans, nil
}

func PopulateExecutionPlan(conn *connection.PGSQLConnection, instanceEntity *integration.Entity) {
	individualQueries, err := ExecutionPlan(conn)
	if err != nil {
		log.Error("Error fetching individual queries: %v", err)
		return
	}
	if len(individualQueries) == 0 {
		log.Info("No individual queries found.")
		return
	}
	log.Info("Populate individual running: %+v", individualQueries)

	for _, model := range individualQueries {
		metricSet := instanceEntity.NewMetricSet("PostgresIndividualQueries")

		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)

		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				setMetric1(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				setMetric1(metricSet, metricName, field.Interface(), sourceType)
			}
		}
	}
}

func setMetric1(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		metricSet.SetMetric(name, value, metric.GAUGE)
	case `attribute`:
		metricSet.SetMetric(name, value, metric.ATTRIBUTE)
	default:
		metricSet.SetMetric(name, value, metric.GAUGE)
	}
}
