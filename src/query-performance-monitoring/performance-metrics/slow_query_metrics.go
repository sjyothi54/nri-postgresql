package performancemetrics

import (
	"fmt"
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	versionSpecificSlowQuery := global_variables.SlowQuery
	if versionSpecificSlowQuery == "" {
		log.Error("Unsupported postgres version")
		return nil, nil, commonutils.ErrUnsupportedVersion
	}
	var query = fmt.Sprintf(versionSpecificSlowQuery, global_variables.DatabaseString, min(global_variables.Args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQueryMetrics
		if scanErr := rows.StructScan(&slowQuery); scanErr != nil {
			return nil, nil, err
		}
		slowQueryMetricsList = append(slowQueryMetricsList, slowQuery)
		slowQueryMetricsListInterface = append(slowQueryMetricsListInterface, slowQuery)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, nil, closeErr
	}
	return slowQueryMetricsList, slowQueryMetricsListInterface, nil
}

func PopulateSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration) []datamodels.SlowRunningQueryMetrics {
	isEligible, err := validations.CheckSlowQueryMetricsFetchEligibility(conn, global_variables.Version)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return nil
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := GetSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return nil
	}
	commonutils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration)
	return slowQueryMetricsList
}
