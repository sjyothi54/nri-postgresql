package validations

import (
	//"github.com/newrelic/infra-integrations-sdk/v4/log"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
)

func CheckPgStatStatementsExtensionEnabled(conn *performance_db_connection.PGSQLConnection) (bool, error) {
	rows, err := conn.Queryx("SELECT count(*) FROM pg_extension WHERE extname = 'pg_stat_statements'")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()
	var count int
	if !rows.Next() {
		return false, nil
	}
	if err := rows.Scan(&count); err != nil {
		log.Error("Error scanning row: ", err.Error())
		return false, err
	}
	return count > 0, nil
}

func CheckPgWaitExtensionEnabled(conn *performance_db_connection.PGSQLConnection) (bool, error) {
	rows, err := conn.Queryx("SELECT count(*) FROM pg_extension WHERE extname = 'pg_wait_sampling'")
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()
	var count int
	if !rows.Next() {
		return false, nil
	}
	if err := rows.Scan(&count); err != nil {
		log.Error("Error scanning row: ", err.Error())
		return false, err
	}
	return count > 0, nil
}
