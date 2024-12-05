package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/query_metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList) {
	connectionInfo := performance_db_connection.DefaultConnectionInfo(&args)
	conn, err := connectionInfo.NewConnection(args.Database)
	if err != nil {
		fmt.Print("Error in connection")
		return
	}
	//queryIdList, err := query_metrics.PopulateSlowRunningMetrics(instanceEntity, conn, args)
	//if err != nil {
	//	fmt.Printf("Error in fetching slow running metrics: %v", err)
	//	return
	//}
	var queryPlanMetrics []datamodels.QueryPlanMetrics
	err = query_metrics.PopulateQueryExecutionMetrics(queryPlanMetrics, instanceEntity, conn, args)
	if err != nil {
		fmt.Print("Error in fetching execution plan metrics check2:", err)
		return
	}

	//err = query_metrics.PopulateWaitEventMetrics(instanceEntity, conn, args)
	//if err != nil {
	//	fmt.Printf("Error in fetching wait event metrics: %v", err)
	//	return
	//}
	//
	//err = query_metrics.PopulateBlockingSessionMetrics(instanceEntity, conn, args)
	//if err != nil {
	//	fmt.Printf("Error in fetching blocking session metrics: %v", err)
	//	return
	//}
}
