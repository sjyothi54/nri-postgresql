package performancemetrics

import (
	"fmt"

	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) {
	isEligible, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, cp.Version)
	if enableCheckError != nil {
		log.Error("Error executing query for eligibility check in PopulateBlockingMetrics: %v", enableCheckError)
		return
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := getBlockingMetrics(conn, cp)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	err := commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting blocking queries: %v", err)
		return
	}
	log.Debug("Successfully ingested blocking metrics ")
}

func getBlockingMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	log.Debug("Executing query to fetch blocking metrics")
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query, error: %v", err)
		return nil, commonutils.ErrUnExpectedError
	}
	defer rows.Close()
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			log.Error("Error scanning row into BlockingSessionMetrics: %v", scanError)
			return nil, scanError
		}
		// For PostgreSQL versions 13 and 12, anonymization of queries does not occur for blocking sessions, so it's necessary to explicitly anonymize them.
		if cp.Version == commonutils.PostgresVersion13 || cp.Version == commonutils.PostgresVersion12 {
			*blockingQueryMetric.BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockedQuery)
			*blockingQueryMetric.BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockingQuery)
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}
	log.Debug("Fetched %d blocking metrics", len(blockingQueriesMetricsList))
	return blockingQueriesMetricsList, nil
}
