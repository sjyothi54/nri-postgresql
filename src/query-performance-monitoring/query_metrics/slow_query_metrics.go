package query_metrics

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.SlowRunningQuery, []int64, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var queryIdList []int64
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
		queryIdList = append(queryIdList, *slowQuery.QueryID)
		log.Info("Slow Query: %+v", slowQuery)
	}

	return slowQueries, queryIdList, nil
}

func GetExplainPlanForSlowQueries(conn *performance_db_connection.PGSQLConnection, queryIdList []int64) (map[int64]string, error) {
	explainPlans := make(map[int64]string)

	for _, queryId := range queryIdList {
		explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) SELECT * FROM pg_stat_statements WHERE queryid = %d", queryId)
		rows, err := conn.Queryx(explainQuery)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var explainResult string
		for rows.Next() {
			var row string
			if err := rows.Scan(&row); err != nil {
				return nil, err
			}
			explainResult += row + "\n"
		}

		explainPlans[queryId] = explainResult
	}

	return explainPlans, nil
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
		log.Error("Error fetching slow-running queries: %v", err)
		return nil, err
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil, errors.New("No slow-running queries found.")
	}
	log.Info("Populate-slow running: %+v", slowQueries)

	explainPlans, err := GetExplainPlanForSlowQueries(conn, queryIdList)
	if err != nil {
		log.Error("Error fetching explain plans: %v", err)
		return nil, err
	}

	for _, model := range slowQueries {
		metricSet := common_utils.CreateMetricSet(instanceEntity, "PostgreSQLQueryExplainGo", args)
		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)
		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}

		log.Info("Metrics set for slow query: %d in database: %s", *model.QueryID, *model.DatabaseName)
		log.Info("Explain plan for query %d: %s", *model.QueryID, explainPlans[*model.QueryID])
	}

	log.Info("Final slow queries: %+v", slowQueries)
	log.Info("Final query IDs: %+v", queryIdList)

	return queryIdList, nil
}
