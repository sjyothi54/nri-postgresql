package query_metrics

import (
	"encoding/json"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
)

func PopulateQueryExecutionMetrics(queryPlanMetrics []datamodels.QueryPlanMetrics, instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) error {
	for _, queryPlanMetric := range queryPlanMetrics {
		fmt.Printf("QueryPlanMetricsssssss: %+v\n\n", queryPlanMetric)
		//query := "EXPLAIN (FORMAT JSON) " + *queryPlanMetric.QueryText
		rows, err := conn.Queryx("")
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

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}
		firstJson, err := json.Marshal(execPlan[0]["Plan"])
		if err != nil {
			log.Error("Failed to marshal firstJson: %v", err)
			continue
		}

		var execPlanMetrics datamodels.QueryExecutionPlanMetrics
		err = json.Unmarshal(firstJson, &execPlanMetrics)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return nil
		}

		fmt.Printf("QueryExecutionPlanMetricsssssss: %+v\n", execPlanMetrics)

		//fmt.Println("mappppppppp", firstJson["Plan"])

		common_utils.SetMetricsParser(instanceEntity, "PostgresqlExecutionPlanMetricsV2", args, execPlanMetrics)

	}
	return nil
}
