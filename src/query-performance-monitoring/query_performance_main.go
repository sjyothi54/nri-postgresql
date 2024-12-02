package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList) error {
	//connectionInfo := performance_db_connection.DefaultConnectionInfo(&args)
	//conn, err := connectionInfo.NewConnection(args.Database)
	//if err != nil {
	//	fmt.Println("Error in connection")
	//}
	fmt.Println("am here")
	metricSet2 := instanceEntity.NewMetricSet(
		"testingV2",
		attribute.Attr("hostname", "12"),
		attribute.Attr("port", "22"),
	)
	err := metricSet2.SetMetric("testMetric2", 5, metric.GAUGE)
	if err != nil {
		fmt.Println("errorr")
		return err
	}
	//queryIdList, err := query_metrics.PopulateSlowRunningMetrics(instanceEntity, conn)
	//if err != nil {
	//	return
	//}
	//fmt.Println("Query ID List: ", queryIdList)
	return nil
}
