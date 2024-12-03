package query_results

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

func ExecutionPlanQuery(conn *connection.PGSQLConnection) error {
	// Prepare the test query
	prepareQuery := "PREPARE test AS SELECT pg_sleep($1)"
	_, err := conn.Queryx(prepareQuery)
	if err != nil {
		log.Error("Error preparing query: %v", err)
		return fmt.Errorf("error preparing query: %w", err)
	}
	log.Info("Query prepared successfully")

	// Print the output of pg_prepared_statements
	rows, err := conn.Queryx("SELECT * FROM pg_prepared_statements")
	if err != nil {
		log.Info("Error querying pg_prepared_statements: %v", err)
		return fmt.Errorf("error querying pg_prepared_statements: %w", err)
	}
	defer rows.Close()
	log.Info("Query executed successfully")

	for rows.Next() {
		var name, statement, prepareTime, parameterTypes, fromSQL string
		if err := rows.Scan(&name, &statement, &prepareTime, &parameterTypes, &fromSQL); err != nil {
			log.Info("Error scanning row: %v", err)
			return fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("Name: %s, Statement: %s, Prepare Time: %s, Parameter Types: %s, From SQL: %s\n",
			name, statement, prepareTime, parameterTypes, fromSQL)
	}

	if err := rows.Err(); err != nil {
		log.Info("Error during row iteration: %v", err)
		return fmt.Errorf("error during row iteration: %w", err)
	}

	log.Info("Rows processed successfully")
	return nil
}
