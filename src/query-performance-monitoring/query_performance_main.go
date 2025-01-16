package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/metrics"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration, databaseList collection.DatabaseList) {

	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		log.Debug("Error creating connection: ", err)
		return
	}

	version, versionErr := metrics.CollectVersion(newConnection)
	versionInt := version.Major
	if versionErr != nil {
		log.Debug("Error fetching version: ", versionErr)
		return
	}
	loadGlobalVariables(args, versionInt, databaseList)

	start := time.Now()
	log.Debug("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration)
	log.Debug("PopulateSlowRunningMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateWaitEventMetrics at ", start)
	_ = performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration)
	log.Debug("PopulateWaitEventMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateBlockingMetrics at ", start)
	_ = performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration)
	log.Debug("PopulateBlockingMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration)
	log.Debug("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration)
	log.Debug("PopulateExecutionPlanMetrics completed in ", time.Since(start))

	log.Debug("Query analysis completed.")
}

func loadGlobalVariables(args args.ArgumentList, version uint64, databaseList collection.DatabaseList) {
	global_variables.Args = args
	global_variables.Version = version
	global_variables.DatabaseString = commonutils.GetDatabaseListInString(databaseList)

	slowQuery, slowQueryErr := commonutils.FetchVersionSpecificSlowQueries(version)
	if slowQueryErr != nil {
		log.Debug("Error fetching slow queries: ", slowQueryErr)
		global_variables.SlowQuery = ""
	} else {
		global_variables.SlowQuery = slowQuery
	}

	blockingQuery, blockingQueryErr := commonutils.FetchVersionSpecificBlockingQueries(version)
	if blockingQueryErr != nil {
		log.Debug("Error fetching blocking queries: ", blockingQueryErr)
		global_variables.BlockingQuery = ""
	} else {
		global_variables.BlockingQuery = blockingQuery
	}

	individualQuery, individualQueryErr := commonutils.FetchVersionSpecificIndividualQueries(version)
	if individualQueryErr != nil {
		log.Debug("Error fetching individual queries: ", individualQueryErr)
		global_variables.IndividualQuery = ""
	} else {
		global_variables.IndividualQuery = individualQuery
	}
}
