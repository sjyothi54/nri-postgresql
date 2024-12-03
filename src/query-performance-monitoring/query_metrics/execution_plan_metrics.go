package query_metrics

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
)

func ExecutionPlanMetrics(conn *performance_db_connection.PGSQLConnection, slowQueriesList []datamodels.SlowRunningQuery) error {

	rows, err := conn.Queryx("select name,parameter_types from pg_prepared_statements;")
	if err != nil {
		fmt.Println("Error in executing prepared statement")
		return err
	}

	fmt.Println("Query Variable testtt: ", rows)

	for rows.Next() {
		var name, statement, prepareTime, parameterTypes, fromSQL string
		if err := rows.Scan(&name, &statement, &prepareTime, &parameterTypes, &fromSQL); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("Name: %s, Statement: %s, Prepare Time: %s, Parameter Types: %s, From SQL: %s\n",
			name, statement, prepareTime, parameterTypes, fromSQL)
	}

	fmt.Println("Query Variable: ", rows)

	//for i, slowQueryMetric := range slowQueriesList {
	//	fmt.Print("Slow Query ", i, ": ", slowQueryMetric)
	//	queryText := slowQueryMetric.QueryText
	//	fmt.Println("Query Text: ", *queryText)
	//
	//	executePrepareStatement := "Prepare test as select * from actor"
	//	fmt.Printf(executePrepareStatement)
	//	_, err := conn.Queryx(executePrepareStatement)
	//	if err != nil {
	//		fmt.Println("Error in executing prepare")
	//		return
	//	}
	//	fmt.Println("eeeeeeeeee")
	//	rows, err := conn.Queryx("select * from pg_prepared_statements")
	//	if err != nil {
	//		fmt.Println("Error in executing prepared statement")
	//	}
	//
	//	fmt.Println("Query Variable: ", rows)
	//	for rows.Next() {
	//		fmt.Println("Row: ", rows)
	//		var parameterData datamodels.Execution_plan_perform_data
	//		if err := rows.StructScan(&parameterData); err != nil {
	//			fmt.Println("Error in scanning row")
	//			continue
	//		}
	//		fmt.Println("parameterData", parameterData)
	//	}
	//
	//	//fmt.Println("Query ID: ", slowQueryMetric.QueryId)
	//	//fmt.Println("Query Text: ", slowQueryMetric.QueryText)
	//	//fmt.Println("Execution Plan: ", slowQueryMetric.QueryPlan)
	//}
	// This function is used to fetch the execution plan metrics
	return nil
}
