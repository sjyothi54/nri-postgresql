package query_performance_monitoring

import (
	_ "fmt"

	_ "github.com/jmoiron/sqlx"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/query_metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

// QueryPerformanceMain is the main function for query performance monitoring.
func QueryPerformanceMainV2(instanceEntity *integration.Entity, cmdArgs args.ArgumentList) {
	// Establish a new database connection
	connectionInfo := connection.DefaultConnectionInfo(&cmdArgs)
	conn, err := connectionInfo.NewConnection(cmdArgs.Database)
	if err != nil {
		log.Error("Error establishing database connection: %v", err)
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Warn("Error closing database connection: %v", cerr)
		}
	}()

	// Check if the pg_wait_sampling extension is enabled
	isExtensionEnabled, err := validations.CheckPgWaitExtensionEnabled(conn.DB)
	if err != nil {
		log.Error("Error checking pg_wait_sampling extension: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return
	}
	log.Info("Extension 'pg_wait_sampling' is enabled.")

	// Populate Wait Event Metrics
	err = query_metrics.PopulateWaitEventMetricsV2(instanceEntity, conn.DB, cmdArgs)
	if err != nil {
		log.Error("Error populating wait event metrics: %v", err)
		return
	}

	// Add additional metric population functions here as needed
	// Example:
	// err = query_metrics.PopulateAnotherMetric(instanceEntity, conn.DB, cmdArgs)
	// if err != nil {
	//     log.Error("Error populating another metric: %v", err)
	//     return
	// }
}
