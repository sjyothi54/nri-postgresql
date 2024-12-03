package query_metrics

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func ExecutionPlanMetrics(slowQueriesList []datamodels.SlowRunningQuery) {

	for i, slowQueryMetric := range slowQueriesList {
		fmt.Print("Slow Query ", i, ": ", slowQueryMetric)
		queryText := slowQueryMetric.QueryText
		fmt.Println("Query Text: ", *queryText)

		executePrepareStatement := "Prepare test as " + *queryText
		fmt.Printf(executePrepareStatement)
		//fmt.Println("Query ID: ", slowQueryMetric.QueryId)
		//fmt.Println("Query Text: ", slowQueryMetric.QueryText)
		//fmt.Println("Execution Plan: ", slowQueryMetric.QueryPlan)
	}
	// This function is used to fetch the execution plan metrics
}
