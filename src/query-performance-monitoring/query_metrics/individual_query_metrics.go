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

func PopulateIndividualMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList, queryIDList []*int64) ([]datamodels.QueryPlanMetrics, error) {
	if len(queryIDList) == 0 {
		log.Warn("queryIDList is empty")
		return nil, nil
	}
	// Building the placeholder string for the IN clause
	query := "SELECT queryId, query FROM pg_stat_monitor WHERE query like 'select * from actor%' and queryId IN ("

	// Convert each queryId to a string and join them with commas
	var idStrings []string
	//for _, id := range queryIDList {
	//	if id != nil {
	//		idStrings = append(idStrings, fmt.Sprintf("%d", *id))
	//	}
	//}

	// Finalize the query string
	query += strings.Join(idStrings, ", ") + ")"

	rows, err := conn.Queryx("SELECT queryId, query FROM pg_stat_monitor WHERE query like 'select * from actor%' ")
	if err != nil {
		fmt.Errorf("Error executing query: %v", err)
		return nil, err
	}
	var individualQueryMetricList []datamodels.QueryPlanMetrics
	defer rows.Close()
	for rows.Next() {
		var individualQueryMetric datamodels.QueryPlanMetrics
		if err := rows.StructScan(&individualQueryMetric); err != nil {
			log.Error("Failed to scan query metrics row: %v", err)
			return nil, err
		}
		individualQueryMetricList = append(individualQueryMetricList, individualQueryMetric)
	}

	fmt.Println("PostgresqlIndividualMetricsV1PostgresqlIndividualMetricsV1", individualQueryMetricList)

	for _, model := range individualQueryMetricList {
		common_utils.SetMetricsParser(instanceEntity, "PostgresqlIndividualMetricsV1", args, model)
	}

	return individualQueryMetricList, nil
}
