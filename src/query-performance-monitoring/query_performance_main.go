package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/query_metrics"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList) {
	connectionInfo := performance_db_connection.DefaultConnectionInfo(&args)
	conn, err := connectionInfo.NewConnection(args.Database)
	if err != nil {
		fmt.Println("Error in connection")
	}
	queryIdList, err := query_metrics.PopulateSlowRunningMetrics(instanceEntity, conn, args)
	if err != nil {
		return
	}
	fmt.Println("Query ID List: ", queryIdList)
}