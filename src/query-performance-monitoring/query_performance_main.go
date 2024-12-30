package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	// newConnection, err := performancedbconnection.OpenDB(args, connectionInfo.DatabaseName())
	if err != nil {
		log.Info("Error creating connection: ", err)
		return
	}
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args)
	performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, args)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	log.Info("Query analysis completed.")
}
