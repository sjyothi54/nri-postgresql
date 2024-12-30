package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList) {
	isPgStatStatementEnabled, enableCheckError := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if enableCheckError != nil {
		log.Debug("Error executing query: %v in PopulateBlockingMetrics", enableCheckError)
		return
	}
	if !isPgStatStatementEnabled {
		log.Debug("Extension 'pg_stat_statements' is not enabled for the database.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := GetBlockingMetrics(conn, args)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, args)
}

func GetBlockingMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	var query = fmt.Sprintf(queries.BlockingQueries, args.QueryCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, err
	}

	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, err
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, closeErr
	}
	return blockingQueriesMetricsList, nil
}
