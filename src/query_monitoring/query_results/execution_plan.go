package query_results

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/connection"
)

func ExecutionPlanQuery(conn *connection.PGSQLConnection) error {
	// Prepare the test query
	rows, err := conn.Queryx("PREPARE test AS SELECT pg_sleep($1)")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return err
	}
	defer rows.Close()
	rows1, err := conn.Queryx("select name from pg_prepared_statements")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return err
	}
	for rows1.Next() {
		var name string
		if err := rows1.Scan(&name); err != nil {
			log.Error("Error scanning row: ", err.Error())
			return err
		}
		log.Info("name: ", name)
	}
	defer rows1.Close()
	return nil
}
