package query_results

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query_monitoring/datamodels"
)

func ExecutionPlanQuery(conn *connection.PGSQLConnection, slowQueries []datamodels.SlowRunningQuery) ([]datamodels.QueryExecutionPlan, error) {
	var executionPlans []datamodels.QueryExecutionPlan

	for i, slowQuery := range slowQueries {
		queryText := slowQuery.QueryText
		fmt.Println("Query Text: ", queryText)
		stmtName := fmt.Sprintf("stmt_%d", i)
		fmt.Println("Statement Name: ", stmtName)
		prepareQuery := fmt.Sprintf("PREPARE %s AS %s", stmtName, *queryText)
		_, err := conn.Queryx(prepareQuery)
		if err != nil {
			return nil, fmt.Errorf("error preparing query: %w", err)
		}

		explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) EXECUTE %s", stmtName)
		rows, err := conn.Queryx(explainQuery)
		if err != nil {
			return nil, fmt.Errorf("error executing explain query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var executionPlan datamodels.QueryExecutionPlan
			if err := rows.StructScan(&executionPlan); err != nil {
				return nil, fmt.Errorf("error scanning execution plan: %w", err)
			}
			executionPlans = append(executionPlans, executionPlan)
			log.Info("Execution Plan: %+v", executionPlan)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over rows: %w", err)
		}

		_, err = conn.Queryx(fmt.Sprintf("DEALLOCATE %s", stmtName))
		if err != nil {
			return nil, fmt.Errorf("error deallocating prepared statement: %w", err)
		}
	}

	return executionPlans, nil
}
