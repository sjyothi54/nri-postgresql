package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
)

func ExecutionPlan(conn *connection.PGSQLConnection) ([]datamodels.ExecutionPlan, error) {
	var query = queries.ExecutionPlanQuery
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var executionPlans []datamodels.ExecutionPlan
	for rows.Next() {
		var executionPlan datamodels.ExecutionPlan
		if err := rows.Scan(&executionPlan.Query, &executionPlan.QueryID); err != nil {
			log.Error("Error scanning row: %v", err)
			return nil, err
		}
		log.Info("query: %s queryid: %s", executionPlan.Query, executionPlan.QueryID)
		executionPlans = append(executionPlans, executionPlan)
	}
	return executionPlans, nil
}
