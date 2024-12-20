package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performanceDbConnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	dbSpecificMap := validations.GetExtensionEnabledDbList(newConnection, args)
	log.Info("dbSpecificMap", dbSpecificMap)
	slowRunningQueries := performance_metrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args)
	performance_metrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args)
	performance_metrics.PopulateBlockingMetrics(newConnection, pgIntegration, args)
	individualQueries := performance_metrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args)
	performance_metrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	fmt.Println("Query analysis completed.")
}
