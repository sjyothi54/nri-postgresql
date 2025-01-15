package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"github.com/newrelic/go-agent/v3/newrelic"
	common_package "github.com/newrelic/nri-postgresql/common-package"
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

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration, databaseList collection.DatabaseList, app *newrelic.Application) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	databaseStringList := commonutils.GetDatabaseListInString(databaseList)
	if len(databaseList) == 0 {
		log.Debug("No databases found")
		return
	}

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

	start := time.Now()
	txn := app.StartTransaction("slow_queries_metrics_go")
	defer txn.End()
	common_package.Txn = txn
	log.Debug("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args, databaseStringList, versionInt, app)
	log.Debug("PopulateSlowRunningMetrics completed in ", time.Since(start))

	waitTxn := app.StartTransaction("wait_queries_metrics_go")
	defer waitTxn.End()
	common_package.Txn = waitTxn
	start = time.Now()
	log.Debug("Starting PopulateWaitEventMetrics at ", start)
	performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args, databaseStringList, app)
	log.Debug("PopulateWaitEventMetrics completed in ", time.Since(start))

	blockingEventsTxn := app.StartTransaction("blocking_queries_go")
	defer blockingEventsTxn.End()
	common_package.Txn = blockingEventsTxn
	start = time.Now()
	log.Debug("Starting PopulateBlockingMetrics at ", start)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, args, databaseStringList, versionInt, app)
	log.Debug("PopulateBlockingMetrics completed in ", time.Since(start))

	individualTxn := app.StartTransaction("individual_txns_go")
	defer individualTxn.End()
	common_package.Txn = individualTxn
	start = time.Now()
	log.Debug("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args, databaseStringList, versionInt, app)
	log.Debug("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	execPlanTxn := app.StartTransaction("execution_plan_go")
	defer execPlanTxn.End()
	common_package.Txn = individualTxn
	start = time.Now()
	log.Debug("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args, app)
	log.Debug("PopulateExecutionPlanMetrics completed in ", time.Since(start))
	log.Debug("Query analysis completed.")
}
