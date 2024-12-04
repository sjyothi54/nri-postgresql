package query_metrics

import (
	"encoding/json"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
)

func PopulateQueryExecutionMetrics(queryPlanMetrics []datamodels.QueryPlanMetrics, instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) error {
	for _, queryPlanMetric := range queryPlanMetrics {
		query := "EXPLAIN (FORMAT JSON) " + queryPlanMetric.QueryText
		rows, err := conn.Queryx(query)
		if err != nil {
			continue
		}
		defer rows.Close()
		if !rows.Next() {
			return nil
		}
		var execPlanJSON string
		if err := rows.Scan(&execPlanJSON); err != nil {
			log.Error("Error scanning row: ", err.Error())
			continue
		}
		//fmt.Print("execPlanJSON:", execPlanJSON)

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}
		firstJson := execPlan[0]

		fmt.Println("mappppppppp", firstJson)

		//common_utils.SetMetricsParser(instanceEntity, "PostgresqlExecutionPlanMetricsV2", args, firstJson)

	}
	return nil
}
