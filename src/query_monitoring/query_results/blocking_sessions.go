
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

func GetBlockingSessionsMetrics(conn *connection.PGSQLConnection) ([]datamodels.BlockingQuery, error) {
	var blockingQueries []datamodels.BlockingQuery
	var query = queries.BlockingQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blockingQuery datamodels.BlockingQuery
		if err := rows.StructScan(&blockingQuery); err != nil {
			return nil, err
		}
		blockingQueries = append(blockingQueries, blockingQuery)
	}

	for _, query := range blockingQueries {
		log.Info("Blocking Query: %+v", query)
	}
	return blockingQueries, nil
}

// PopulateBlockingSessionsMetrics fetches blocking sessions metrics and populates them into the metric set
func PopulateBlockingSessionsMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' enabled.")
		blockingQueries, err := GetBlockingSessionsMetrics(conn)
		if err != nil {
			log.Error("Error fetching blocking queries: %v", err)
			return
		}

		if len(blockingQueries) == 0 {
			log.Info("No blocking queries found.")
			return
		}
		log.Info("Populate blocking queries: %+v", blockingQueries)

		for _, model := range blockingQueries {
			metricSet := instanceEntity.NewMetricSet("PostgresBlockingQueriesGoV1")
			log.Info("Creating metric set for blocking query: %+v", model)

			modelValue := reflect.ValueOf(model)
			modelType := reflect.TypeOf(model)

			for i := 0; i < modelValue.NumField(); i++ {
				field := modelValue.Field(i)
				fieldType := modelType.Field(i)
				metricName := fieldType.Tag.Get("metric_name")
				sourceType := fieldType.Tag.Get("source_type")

				if field.Kind() == reflect.Ptr && !field.IsNil() {
					setMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
				} else if field.Kind() != reflect.Ptr {
					setMetric(metricSet, metricName, field.Interface(), sourceType)
				}
				log.Info("Metric set created: %+v", metricSet)
			}
		}
	} else {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}
}