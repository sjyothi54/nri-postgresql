package performancemetrics

import (
	"fmt"
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration) error {
	isEligible, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(conn, global_variables.Version)
	if enableCheckError != nil {
		log.Debug("Error executing query: %v in PopulateBlockingMetrics", enableCheckError)
		return commonutils.ErrUnExpectedError
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return commonutils.ErrNotEligible
	}
	blockingQueriesMetricsList, blockQueryFetchErr := GetBlockingMetrics(conn)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return commonutils.ErrUnExpectedError
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return nil
	}
	commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration)
	return nil
}

func GetBlockingMetrics(conn *performancedbconnection.PGSQLConnection) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery := global_variables.BlockingQuery
	if versionSpecificBlockingQuery == "" {
		log.Error("Unsupported postgres version")
		return nil, commonutils.ErrUnsupportedVersion
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, global_variables.DatabaseString, min(global_variables.Args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, err
	}
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		if global_variables.Version == commonutils.PostgresVersion13 || global_variables.Version == commonutils.PostgresVersion12 {
			*blockingQueryMetric.BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockedQuery)
			*blockingQueryMetric.BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockingQuery)
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, closeErr
	}
	return blockingQueriesMetricsList, nil
}
