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

func getBlockingSessionMetrics(conn *performance_db_connection.PGSQLConnection) ([]interface{}, error) {
	var blockingSessionMetrics []interface{}
	var query = queries.BlockingQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var blockingSessionMetric datamodels.BlockingQuery
		if err := rows.StructScan(&blockingSessionMetric); err != nil {
			return nil, err
		}
		blockingSessionMetrics = append(blockingSessionMetrics, blockingSessionMetric)
	}

	return blockingSessionMetrics, nil
}

func PopulateBlockingSessionMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList, pgIntegration *integration.Integration) error {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return err
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return errors.New("extension 'pg_wait_sampling' is not enabled")
	}
	log.Info("Extension 'pg_wait_sampling' enabled.")
	blockingSessionMetrics, err := getBlockingSessionMetrics(conn)
	if err != nil {
		log.Error("Error fetching blocking-session metrics: %v", err)
		return err
	}

	if len(blockingSessionMetrics) == 0 {
		log.Info("No blocking-session metrics found.")
		return errors.New("no blocking-session metrics found")
	}

	log.Info("blockingSessionMetrics %+v", blockingSessionMetrics)

	common_utils.SetMetricsParser(instanceEntity, "PostgresqlBlockingSessionSample", args, pgIntegration, blockingSessionMetrics)

	return nil
}
