package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
)

func ExecutionPlan(conn *connection.PGSQLConnection) {
	query := queries.ExecutionPlanQuery
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var queryID string
		var queryText string
		if err := rows.Scan(&queryID, &queryText); err != nil {
			log.Error("Error scanning row: %v", err)
			return
		}
		log.Info("Query ID: %s, Query: %s", queryID, queryText)
	}
}
