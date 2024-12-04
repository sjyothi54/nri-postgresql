package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/queries"
)

func ExecutionPlan(conn *connection.PGSQLConnection) {
	slowQueries, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return
	}

	query := queries.ExecutionPlanQuery
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	defer rows.Close()

	queryMap := make(map[string]string)
	for rows.Next() {
		var queryID string
		var queryText string
		if err := rows.Scan(&queryID, &queryText); err != nil {
			log.Error("Error scanning row: %v", err)
			return
		}
		queryMap[queryID] = queryText
	}

	// Fetch query IDs and texts from pg_stat_monitor
	monitorQuery := queries.ExecutionPlanMonitorQuery
	monitorRows, err := conn.Queryx(monitorQuery)
	if err != nil {
		log.Error("Error executing monitor query: %v", err)
		return
	}
	defer monitorRows.Close()

	for monitorRows.Next() {
		var queryID string
		var queryText string
		if err := monitorRows.Scan(&queryID, &queryText); err != nil {
			log.Error("Error scanning monitor row: %v", err)
			return
		}
		queryMap[queryID] = queryText
	}

	// Debug log to print queryMap
	for queryID, queryText := range queryMap {
		log.Debug("QueryMap - Query ID: %s, Query: %s", queryID, queryText)
	}

	for _, slowQuery := range slowQueries {
		// Debug log to print slowQuery.QueryID
		if slowQuery.QueryID == nil {
			log.Debug("SlowQuery - Query ID is nil")
			continue
		}
		log.Debug("SlowQuery - Query ID: %d", *slowQuery.QueryID)
		if queryText, exists := queryMap[*slowQuery.QueryID]; exists {
			log.Info("Matching Query ID: %s, Query: %s", *slowQuery.QueryID, queryText)
		} else {
			log.Info("No matching query found for Query ID: %d", *slowQuery.QueryID)
		}
	}
}
