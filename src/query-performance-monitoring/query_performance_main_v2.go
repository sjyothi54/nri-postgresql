package query_performance_monitoring

//
//import (
//	"github.com/newrelic/infra-integrations-sdk/v3/integration"
//	"github.com/newrelic/infra-integrations-sdk/v3/log"
//
//	"github.com/newrelic/nri-postgresql/src/args"
//	"github.com/newrelic/nri-postgresql/src/connection"
//	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/query_metrics"
//	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
//)
//
//// QueryPerformanceMain orchestrates the collection of query performance metrics.
//func QueryPerformanceMainV2(instanceEntity *integration.Entity, cmdArgs args.ArgumentList) {
//	// Establish a new database connection
//	connectionInfo := connection.DefaultConnectionInfo(&cmdArgs)
//	db, err := connectionInfo.NewConnection(cmdArgs.Database)
//	if err != nil {
//		log.Error("Error establishing database connection: %v", err)
//		return
//	}
//
//	log.Info("Starting query performance monitoring for instance %s:%s", cmdArgs.Hostname, cmdArgs.Port)
//
//	// Check if the pg_wait_sampling extension is enabled
//	isExtensionEnabled, err := validations.CheckPgWaitExtensionEnabled(db)
//	if err != nil {
//		log.Error("Error checking pg_wait_sampling extension: %v", err)
//		return
//	}
//	if !isExtensionEnabled {
//		log.Info("Extension 'pg_wait_sampling' is not enabled.")
//		return
//	}
//	log.Info("Extension 'pg_wait_sampling' is enabled.")
//
//	// Populate Wait Event Metrics
//	err = query_metrics.PopulateWaitEventMetricsV2(instanceEntity, db, cmdArgs)
//	if err != nil {
//		log.Error("Error populating wait event metrics: %v", err)
//		return
//	}
//}
