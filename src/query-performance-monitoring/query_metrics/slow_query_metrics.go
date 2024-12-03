package query_metrics

import (
	"errors"
	"fmt"
	"reflect"

	// "strings"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.SlowRunningQuery, []string, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	//var queryIdList []int64
	var queryTextList []string
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
		//queryIdList = append(queryIdList, *slowQuery.QueryID)
		queryTextList = append(queryTextList, *slowQuery.QueryText)
		log.Info("Slow Query: %+v", slowQuery)
	}

	/*var queryIdListStr []string
	  for _, id := range queryIdList {
	      queryIdListStr = append(queryIdListStr, fmt.Sprintf("%d", id))
	  }*/
	return slowQueries, queryTextList, nil
}

func GetExplainPlanForSlowQueries(conn *performance_db_connection.PGSQLConnection, queryTextList []string) (map[string]string, error) {
	explainPlans := make(map[string]string)

	for idx, queryText := range queryTextList {
		fmt.Printf("Executing for Query Text: %s\n", queryText)

		// Define unique name for prepared statement per query
		planName := fmt.Sprintf("plan%d", idx)

		// Prepare the statement
		prepareQuery := fmt.Sprintf("PREPARE %s AS %s;", planName, queryText)
		fmt.Printf("Preparing Statement: %s\n", prepareQuery)

		if _, err := conn.Queryx(prepareQuery); err != nil {
			fmt.Printf("Error preparing statement: %s, %v\n", planName, err)
			continue // Log and continue to next statement on error
		}

		// Execute the prepared statement
		explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) EXECUTE %s;", planName)
		fmt.Printf("Executing Query: %s\n", explainQuery)

		rows, err := conn.Queryx(explainQuery)
		if err != nil {
			fmt.Printf("Error executing prepared statement: %s, %v\n", planName, err)
			continue // Proceed to cleanup and next execution if failure
		}

		var explainResult string
		for rows.Next() {
			var row string
			if err := rows.Scan(&row); err != nil {
				fmt.Printf("Error fetching row: %v\n", err)
				continue // Continue on row scan error
			}
			explainResult += row + "\n"
		}
		rows.Close() // Ensure all resources are released

		explainPlans[queryText] = explainResult

		// Deallocate the prepared statement to avoid session pollution
		deallocQuery := fmt.Sprintf("DEALLOCATE %s;", planName)
		fmt.Printf("Deallocating Statement: %s\n", deallocQuery)

		if _, err := conn.Queryx(deallocQuery); err != nil {
			fmt.Printf("Error deallocating statement: %s, %v\n", planName, err)
		}
	}

	return explainPlans, nil
}

func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) ([]string, error) {
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
	slowQueries, queryTextList, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil, err
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil, errors.New("No slow-running queries found.")
	}
	log.Info("Populate-slow running: %+v", slowQueries)

	explainPlans, err := GetExplainPlanForSlowQueries(conn, queryTextList)
	if err != nil {
		log.Error("Error fetching explain plans: %v", err)
		return nil, err
	}

	for _, model := range explainPlans {
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

		log.Info("Metrics set for slow query text: %s ", explainPlans[model])
		//log.Info("Explain plan for query %d: %s", *model.QueryText)
	}

	log.Info("Final slow queries: %+v", slowQueries)
	log.Info("Final query IDs: %+v", queryTextList)

	return queryTextList, nil
}
