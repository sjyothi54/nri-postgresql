package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetBlockingMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	var query = fmt.Sprintf(queries.BlockingQueries, args.QueryCountThreshold)
	log.Info("Blocking query :", query)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if err := rows.StructScan(&blockingQueryMetric); err != nil {
			return nil, err
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	return blockingQueriesMetricsList, nil
}

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return
	}
	log.Info("Extension 'pg_stat_statements' enabled.")
	blockingQueriesMetricsList, err := GetBlockingMetrics(conn, args)
	if err != nil {
		log.Error("Error fetching Blocking queries: %v", err)
		return
	}

	if len(blockingQueriesMetricsList) == 0 {
		log.Info("No Blocking queries found.")
		return
	}
	commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, args)

}
