package query_results

import (
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/validations"
)

func GetIndividualMetrics(conn *connection.PGSQLConnection) ([]datamodels.IndividualQuery, error) {
	var individualQueries []datamodels.IndividualQuery
	var query = queries.IndividualQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var individualQuery datamodels.IndividualQuery
		if err := rows.StructScan(&individualQuery); err != nil {
			return nil, err
		}
		individualQueries = append(individualQueries, individualQuery)
	}

	for _, query := range individualQueries {
		log.Info("Individual Query: %+v", query)
	}
	return individualQueries, nil
}

// PopulateSlowRunningMetrics fetches slow-running metrics and populates them into the metric set
func PopulateIndividualMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	isExtensionEnabled, err := validations.CheckPgStatMonitorExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_stat_monitor' enabled.")
		individualQueries, err := GetIndividualMetrics(conn)
		if err != nil {
			log.Error("Error fetching individual queries: %v", err)
			return
		}

		if len(individualQueries) == 0 {
			log.Info("No individual queries found.")
			return
		}
		log.Info("Populate individualrunning: %+v", individualQueries)

		for _, model := range individualQueries {
			metricSet := instanceEntity.NewMetricSet("PostgresIndividualQueriesGo")

			modelValue := reflect.ValueOf(model)
			modelType := reflect.TypeOf(model)

			for i := 0; i < modelValue.NumField(); i++ {
				field := modelValue.Field(i)
				fieldType := modelType.Field(i)
				metricName := fieldType.Tag.Get("metric_name")
				sourceType := fieldType.Tag.Get("source_type")
				log.Info("Setting metric: %s with value: %v and source type: %s", metricName, field.Interface(), sourceType)
				if field.Kind() == reflect.Ptr && !field.IsNil() {
					SetMetrics(metricSet, metricName, field.Elem().Interface(), sourceType)
				} else if field.Kind() != reflect.Ptr {
					SetMetrics(metricSet, metricName, field.Interface(), sourceType)
				}
			}

			log.Info("Metrics set for slow query: %s", *model.Query)
		}
	} else {
		log.Info("Extension 'pg_stat_monitor' is not enabled.")
		return
	}

}
