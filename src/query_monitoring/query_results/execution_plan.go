package query_results

import (
	"fmt"

	"github.com/newrelic/nri-postgresql/src/connection"
)

func ExecutionPlanQuery(conn *connection.PGSQLConnection) error {
	// Prepare the test query
	prepareQuery := "PREPARE test AS SELECT pg_sleep($1)"
	_, err := conn.Queryx(prepareQuery)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}

	// Print the output of pg_prepared_statements
	rows, err := conn.Queryx("SELECT * FROM pg_prepared_statements;")
	if err != nil {
		return fmt.Errorf("error querying pg_prepared_statements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, statement, prepareTime, parameterTypes, fromSQL string
		if err := rows.Scan(&name, &statement, &prepareTime, &parameterTypes, &fromSQL); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("Name: %s, Statement: %s, Prepare Time: %s, Parameter Types: %s, From SQL: %s\n",
			name, statement, prepareTime, parameterTypes, fromSQL)
	}

	return nil
}
