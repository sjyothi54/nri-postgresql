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

func GetSlowRunningMetrics(connList []*performanceDbConnection.PGSQLConnection) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	var query = queries.SlowQueries
	for _, conn := range connList {
		rows, err := conn.Queryx(query)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var slowQuery datamodels.SlowRunningQueryMetrics
			if err := rows.StructScan(&slowQuery); err != nil {
				return nil, nil, err
			}
			slowQueryMetricsList = append(slowQueryMetricsList, slowQuery)
			slowQueryMetricsListInterface = append(slowQueryMetricsListInterface, slowQuery)
		}
	}

	return slowQueryMetricsList, slowQueryMetricsListInterface, nil
}

func PopulateSlowRunningMetrics(pgIntegration *integration.Integration, args args.ArgumentList) []datamodels.SlowRunningQueryMetrics {
	dbConnList, err := validations.CheckDbsWithSlowQueryMetricsEligibility()
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if len(dbConnList) == 0 {
		log.Info("Extension PopulateSlowRunningMetrics is not eligible. No databases found.")
		return nil
	}

	log.Info("Extension 'pg_stat_statements' enabled.")
	slowQueryMetricsList, slowQueryMetricsListInterface, err := GetSlowRunningMetrics(dbConnList)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Info("No slow-running queries found.")
		return nil
	}
	common_utils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration, args)
	return slowQueryMetricsList

}
