package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetWaitEventMetrics(dbConnList []*performanceDbConnection.PGSQLConnection) ([]interface{}, error) {
	var waitEventMetricsList []interface{}
	var query = queries.WaitEvents
	for _, conn := range dbConnList {
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
	}
	return waitEventMetricsList, nil
}

func PopulateWaitEventMetrics(pgIntegration *integration.Integration, args args.ArgumentList) {
	dbConnList, err := validations.CheckDbWithWaitMetricsEligibility()
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}

	if len(dbConnList) == 0 {
		log.Info("Extension PopulateWaitEventMetrics is not eligible. No databases found.")
		return
	}

	log.Info("Extension 'pg_wait_sampling' enabled.")
	waitEventMetricsList, err := GetWaitEventMetrics(dbConnList)
	if err != nil {
		log.Error("Error fetching wait event queries: %v", err)
		return
	}

	if len(waitEventMetricsList) == 0 {
		log.Info("No wait event queries found.")
		return
	}

	common_utils.IngestMetric(waitEventMetricsList, "PostgresWaitEvents", pgIntegration, args)

}
