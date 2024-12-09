package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/query_metrics"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList) {
	connectionInfo := performance_db_connection.DefaultConnectionInfo(&args)
	conn, err := connectionInfo.NewConnection(args.Database)
	if err != nil {
		fmt.Print("Error in connection")
		return
	}
	_, err = query_metrics.PopulateSlowRunningMetrics(instanceEntity, conn, args)
	if err != nil {
		fmt.Printf("Error in fetching slow running metrics: %v", err)
		return
	}

	ms := common_utils.CreateMetricSet(instanceEntity, "PostgresqlGoWaitEventMetrics", args)
	err = ms.SetMetric("event_type", "PostgresqlWaitEventMetricsV1Test", metric.ATTRIBUTE)
	if err != nil {
		log.Error("Error setting event_type attribute: %v", err)
		return
	}

	//err = query_metrics.PopulateWaitEventMetrics(instanceEntity, conn, args)
	//if err != nil {
	//	fmt.Print("Error in fetching individual metrics: ", err)
	//	return
	//

	//err = query_metrics.PopulateWaitEventMetrics(instanceEntity, conn, args)
	//if err != nil {
	//	fmt.Printf("Error in fetching wait event metrics: %v\n", err)
	//	return
	//}
	//fmt.Println(queryIdList)
	//_, err = query_metrics.PopulateIndividualMetrics(instanceEntity, conn, args, queryIdList)
	//if err != nil {
	//	fmt.Print("Error in fetching execution plan metrics check2:", err)
	//	return
	//}
	//query_metrics.PopulateQueryExecutionMetrics(individualMetrics, instanceEntity, conn, args)

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
