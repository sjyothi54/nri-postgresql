package query_metrics

import (
	"errors"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func getWaitEventMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.WaitEventQuery, error) {
	var waitEventMetrics []datamodels.WaitEventQuery
	var query = queries.WaitEvents
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var waitEventMetric datamodels.WaitEventQuery
		if err := rows.StructScan(&waitEventMetric); err != nil {
			return nil, err
		}
		waitEventMetrics = append(waitEventMetrics, waitEventMetric)
	}

	return waitEventMetrics, nil
}

func PopulateWaitEventMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) error {
	isExtensionEnabled, err := validations.CheckPgWaitExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return err
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return errors.New("extension 'pg_wait_sampling' is not enabled")
	}
	log.Info("Extension 'pg_wait_sampling' enabled.")
	waitEventMetrics, err := getWaitEventMetrics(conn)
	if err != nil {
		log.Error("Error fetching wait-event metrics: %v", err)
		return err
	}

	if len(waitEventMetrics) == 0 {
		log.Info("No wait-event metrics found.")
		return nil
	}

	log.Info("WaitEventMetrics %+v", waitEventMetrics)

	//ms := common_utils.CreateMetricSet(instanceEntity, "PostgresqlWaitEventMetricsV1", args)
	//err = ms.SetMetric("event_type", "PostgresqlWaitEventMetricsV1", metric.ATTRIBUTE)
	if err != nil {
		log.Error("Error setting event_type attribute: %v", err)
		return err
	}
	for _, model := range waitEventMetrics {
		common_utils.SetMetricsParser(instanceEntity, "PostgresqlWaitEventMetricsV1", args, model)
	}

	return nil
}
