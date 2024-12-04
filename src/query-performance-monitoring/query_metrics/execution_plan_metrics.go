package query_metrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"strings"
)

func ExecutionPlanMetrics(conn *performance_db_connection.PGSQLConnection, queryIDList []*int64) error {
	fmt.Println("Query ID List: ", queryIDList)
	if len(queryIDList) == 0 {
		log.Warn("queryIDList is empty")
		return nil
	}
	// Building the placeholder string for the IN clause
	query := "SELECT queryId, query FROM pg_stat_monitor WHERE queryId IN ("

	// Convert each queryId to a string and join them with commas
	var idStrings []string
	for _, id := range queryIDList {
		idStrings = append(idStrings, fmt.Sprintf("%d", *id))
	}

	// Finalize the query string
	query += strings.Join(idStrings, ", ") + ")"

	fmt.Printf("Queryfinallformatted: %s\n", query)

	rows, err := conn.Queryx(query)
	if err != nil {
		fmt.Errorf("Error executing query: %v", err)
		return err
	}
	var metricList []datamodels.QueryPlanMetrics
	defer rows.Close()
	for rows.Next() {
		var metric datamodels.QueryPlanMetrics
		if err := rows.StructScan(&metric); err != nil {
			log.Error("Failed to scan query metrics row: %v", err)
			return err
		}
		metricList = append(metricList, metric)
	}

	fmt.Println(metricList)
	return nil
}
