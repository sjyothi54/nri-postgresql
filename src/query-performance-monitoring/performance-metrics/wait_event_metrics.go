package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) error {
	waitEventMetricsList, waitEventErr := getWaitEventMetrics(conn, cp,enabledExtensions)
	if waitEventErr != nil {
		log.Error("Error fetching wait event queries: %v", waitEventErr)
		return commonutils.ErrUnExpectedError
	}
	if len(waitEventMetricsList) == 0 {
		log.Debug("No wait event queries found.")
		return nil
	}
	err := commonutils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting wait event queries: %v", err)
		return err
	}
	return nil
}

func getWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query string

	// Check if pg_wait_sampling is enabled
	if enabledExtensions["pg_wait_sampling"] && enabledExtensions["pg_stat_statements"] {
		query = fmt.Sprintf(queries.WaitEvents, cp.Databases, cp.QueryMonitoringCountThreshold)
	} else {
		query = fmt.Sprintf(queries.WaitEventsWithoutExtension, cp.Databases, cp.QueryMonitoringCountThreshold)
	}

	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if waitScanErr := rows.StructScan(&waitEvent); waitScanErr != nil {
			return nil, err
		}

		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	return waitEventMetricsList, nil
}
