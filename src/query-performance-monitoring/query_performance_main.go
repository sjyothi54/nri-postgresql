package queryPerformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	start := time.Now()
	log.Info("Start PopulateSlowRunningMetrics:", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateSlowRunningMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateWaitEventMetrics:", start)
	performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateWaitEventMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateBlockingMetrics:", start)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateBlockingMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateIndividualQueryMetrics:", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args)
	log.Info("End PopulateIndividualQueryMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateExecutionPlanMetrics:", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	log.Info("End PopulateExecutionPlanMetrics:", time.Since(start).Seconds())

	log.Info("Query analysis completed.")
}
