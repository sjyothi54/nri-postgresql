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

func PopulateWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList) {
	isExtensionEnabled, err := validations.CheckPgWaitSamplingExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}
	log.Info("Extension 'pg_wait_sampling' enabled.")
	waitEventMetricsList, err := GetWaitEventMetrics(conn, args)
	if err != nil {
		log.Error("Error fetching wait event queries: %v", err)
		return
	}

	if len(waitEventMetricsList) == 0 {
		log.Info("No wait event queries found.")
		return
	}

	commonutils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, args)

}

func GetWaitEventMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query = fmt.Sprintf(queries.WaitEvents, args.QueryCountThreshold)
	log.Info("Wait event query :", query)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var waitEvent datamodels.WaitEventMetrics
		if err := rows.StructScan(&waitEvent); err != nil {
			return nil, err
		}
		waitEventMetricsList = append(waitEventMetricsList, waitEvent)
	}
	return waitEventMetricsList, nil
}
