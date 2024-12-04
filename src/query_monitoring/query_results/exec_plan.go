package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"

	"github.com/newrelic/nri-postgresql/src/connection"
)

func ExecutionPlan(conn *connection.PGSQLConnection) {
	slowQueries, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return
	}

	for _, query := range slowQueries {
		log.Info("Slow Query ID: %s", *query.QueryID)
	}
}
