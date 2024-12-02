package query_metrics

import (
	"errors"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
	"reflect"
)

func GetSlowRunningMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.SlowRunningQuery, []int64, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var qIdList []int64
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
		qIdList = append(qIdList, *slowQuery.QueryID)
	}

	return slowQueries, qIdList, nil
}

func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) ([]int64, error) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil, err
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return nil, errors.New("Extension 'pg_stat_statements' is not enabled.")
	}
	log.Info("Extension 'pg_stat_statements' enabled.")
	slowQueries, queryIdList, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %d", err)
		return nil, err
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil, errors.New("No slow-running queries found.")
	}
	fmt.Println("Slow Queries: ", slowQueries)
	for _, query := range slowQueries {
		log.Info("Populate-slow running: QueryID: %d, QueryText: %s, DatabaseName: %s", *query.QueryID, *query.QueryText, *query.DatabaseName)
	}
	log.Info("Populate-slow running: %+v", slowQueries)
	metricSet2 := instanceEntity.NewMetricSet("PostgresSlowQueriesV2")
	metricSet2.SetMetric("test_metric", 10, metric.GAUGE)
	//common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
	for _, model := range slowQueries {

		fmt.Printf("Model: %v\n", model)

		metricSet := instanceEntity.NewMetricSet("PostgresSlowQueriesV2")

		modelValue := reflect.ValueOf(model)
		fmt.Println("Model Value: ", modelValue)
		modelType := reflect.TypeOf(model)
		fmt.Println("Model Type: ", modelType)
		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				fmt.Println("in ifffff")
				metricSet.SetMetric(metricName, field.Elem().Interface(), metric.GAUGE)
				//common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				fmt.Println("in else ifffff")
				metricSet.SetMetric(metricName, field.Elem().Interface(), metric.GAUGE)
				common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}
		log.Info("Metrics set for slow query: %s in database: %s", *model.QueryID, *model.DatabaseName)
	}

	return queryIdList, nil
}
