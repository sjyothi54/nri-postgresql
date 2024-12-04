package query_metrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
)

func PopulateQueryExecutionMetrics(queryPlanMetrics []datamodels.QueryPlanMetrics, instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) {
	for _, queryPlanMetric := range queryPlanMetrics {
		query := "EXPLAIN (FORMAT JSON) " + queryPlanMetric.QueryText
		rows, err := conn.Queryx(query)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			fmt.Print(rows)
		}
	}
}
