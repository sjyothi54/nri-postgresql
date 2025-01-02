package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		log.Info("Error creating connection: ", err)
		return
	}

	start := time.Now()
	log.Info("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args)
	log.Info("PopulateSlowRunningMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Info("Starting PopulateWaitEventMetrics at ", start)
	performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args)
	log.Info("PopulateWaitEventMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Info("Starting PopulateBlockingMetrics at ", start)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, args)
	log.Info("PopulateBlockingMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Info("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args)
	log.Info("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Info("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	log.Info("PopulateExecutionPlanMetrics completed in ", time.Since(start))

	log.Info("Query analysis completed.")
}
