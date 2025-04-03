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
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) error {
    querySource, eligibleCheckErr := validations.CheckWaitEventMetricsFetchEligibility(enabledExtensions)
    if eligibleCheckErr != nil {
        log.Error("Error checking wait event metrics fetch eligibility: %v", eligibleCheckErr)
        return commonutils.ErrUnExpectedError
    }
    if querySource == "pg_stat_activity" {
        log.Debug("Using 'pg_stat_activity' as the query source due to missing extensions.")
    }
    waitEventMetricsList, waitEventErr := getWaitEventMetrics(conn, cp, enabledExtensions)
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

    // Determine which query to use based on extension eligibility
    querySource, err := validations.CheckWaitEventMetricsFetchEligibility(enabledExtensions)
    if err != nil {
        log.Error("Error checking wait event metrics fetch eligibility: %v", err)
        return nil, err
    }

    var query string
    if querySource == "pg_stat_activity" {
        query = fmt.Sprintf(queries.WaitEventsRds)
        fmt.Println("wait rds query",query)
    } else {
        query = fmt.Sprintf(queries.WaitEvents, cp.Databases, cp.QueryMonitoringCountThreshold)
    }

    rows, err := conn.Queryx(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var waitEvent datamodels.WaitEventMetrics
        if waitScanErr := rows.StructScan(&waitEvent); waitScanErr != nil {
            return nil, waitScanErr
        }
        waitEventMetricsList = append(waitEventMetricsList, waitEvent)
    }
    return waitEventMetricsList, nil
}
