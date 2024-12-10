package query_metrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"strings"
)

func getIndividualMetrics(conn *performance_db_connection.PGSQLConnection, queryIdList []*int64) ([]interface{}, error) {
	var individualQueryMetricList []interface{}
	var individualQuerySearchQuery = getIndividualQueryStatementSearchQuery(queryIdList)

	fmt.Println("individualQuerySearch::::", individualQuerySearchQuery)

	individualQueriesRows, err := conn.Queryx("select query from pg_stat_monitor WHERE query like 'select * from actor%'")

	if err != nil {
		fmt.Printf("Error in fetching individual query metrics: %v", err)
		return nil, err
	}
	for individualQueriesRows.Next() {
		var individualQueryMetric datamodels.QueryPlanMetrics
		if err := individualQueriesRows.StructScan(&individualQueryMetric); err != nil {
			fmt.Printf("Failed to scan query metrics row: %v\n", err)
			return nil, err
		}
		fmt.Println("individualQueryMetric::::", individualQueryMetric)
		individualQueryMetricList = append(individualQueryMetricList, individualQueryMetric)
	}
	return individualQueryMetricList, nil
}

func PopulateIndividualMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList, queryIDList []*int64, pgIntegration *integration.Integration) ([]interface{}, error) {
	if len(queryIDList) == 0 {
		log.Warn("queryIDList is empty")
		return nil, nil
	}

	individualQueriesMetricsList, err := getIndividualMetrics(conn, queryIDList)
	if err != nil {
		return nil, err
	}

	common_utils.SetMetricsParser(instanceEntity, "PostgresqlIndividualMetricsV1", args, pgIntegration, individualQueriesMetricsList)

	return individualQueriesMetricsList, nil
}

func getIndividualQueryStatementSearchQuery(queryIDList []*int64) string {
	query := "SELECT queryId, query FROM pg_stat_monitor WHERE query like 'select * from actor%' and queryId IN ("

	var idStrings []string
	for _, id := range queryIDList {
		if id != nil {
			idStrings = append(idStrings, fmt.Sprintf("%d", *id))
		}
	}

	// Finalize the query string
	query += strings.Join(idStrings, ", ") + ")"

	return query

}
