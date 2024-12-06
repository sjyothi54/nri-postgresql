package query_results

import (
	"reflect"

	// "github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/validations"
)

func GetWaitEventMetrics(conn *connection.PGSQLConnection) ([]datamodels.WaitEventQuery, error) {
	var waitEvents []datamodels.WaitEventQuery
	var query = queries.WaitEvents
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var waitEvent datamodels.WaitEventQuery
		if err := rows.StructScan(&waitEvent); err != nil {
			return nil, err
		}
		waitEvents = append(waitEvents, waitEvent)
	}

	for _, event := range waitEvents {
		log.Info("Wait Event: %+v", event)
	}
	return waitEvents, nil
}

// PopulateWaitEventMetrics fetches wait event metrics and populates them into the metric set
func PopulateWaitEventMetrics(instanceEntity *integration.Entity, conn *connection.PGSQLConnection, query string) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' enabled.")
		waitEvents, err := GetWaitEventMetrics(conn)
		if err != nil {
			log.Error("Error fetching wait event queries: %v", err)
			return
		}

		if len(waitEvents) == 0 {
			log.Info("No wait event queries found.")
			return
		}
		log.Info("Populate wait events: %+v", waitEvents)

		for _, model := range waitEvents {
			metricSet := instanceEntity.NewMetricSet("PostgresWaitEventsGoV1")
			log.Info("Creating metric set for wait event: %+v", model)

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
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}
}