package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
)

func ExecutionPlan(conn *connection.PGSQLConnection) {
	var query = queries.ExecutionPlanQuery
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var executionPlan string
		if err := rows.Scan(&executionPlan); err != nil {
			log.Error("Error scanning row: %v", err)
			return
		}
		log.Info("Execution Plan: %s", executionPlan)
	}
}
