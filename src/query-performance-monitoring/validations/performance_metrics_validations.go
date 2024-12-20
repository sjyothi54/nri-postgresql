package validations

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

const pg_stat_statements_extension = "pg_stat_statements"
const pg_stat_monitor_extension = "pg_stat_monitor"
const pg_wait_sampling_extension = "pg_wait_sampling"

func checkIsExtensionEnabled(conn *performanceDbConnection.PGSQLConnection, extensionName string) (bool, error) {
	rows, err := conn.Queryx(fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", extensionName))
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return false, err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Error("Error scanning rows: ", err.Error())
		}
	}
	if err := rows.Err(); err != nil {
		log.Error(err.Error())
	}

	return count > 0, nil
}

func CheckPgWaitSamplingExtensionEnabled(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return checkIsExtensionEnabled(conn, pg_wait_sampling_extension)
}

func CheckPgStatStatementsExtensionEnabled(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return checkIsExtensionEnabled(conn, pg_stat_statements_extension)
}

func CheckPgStatMonitorExtensionEnabled(conn *performanceDbConnection.PGSQLConnection) (bool, error) {
	return checkIsExtensionEnabled(conn, pg_stat_monitor_extension)
}

func GetExtensionEnabledDbList(conn *performanceDbConnection.PGSQLConnection, args args.ArgumentList) map[string][]string {
	databaseRows, err := conn.Queryx(queries.ListOfDatabases)
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return nil
	}
	var databasesList []string
	defer databaseRows.Close()
	for databaseRows.Next() {
		var dbName string
		if err := databaseRows.Scan(&dbName); err != nil {
			log.Error("Error scanning rows: ", err.Error())
		}
		databasesList = append(databasesList, dbName)
	}
	var extensionDbMap = make(map[string][]string)
	for _, dbName := range databasesList {
		dbConn, err := performanceDbConnection.OpenDB(args, dbName)
		isPgStatStatementExtensionEnabled, err := CheckPgStatStatementsExtensionEnabled(dbConn)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return nil
		}
		if !isPgStatStatementExtensionEnabled {
			log.Info("Extension 'pg_stat_statements' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_stat_statements_extension] = append(extensionDbMap[pg_stat_statements_extension], dbName)

		isPgStatMonitorExtensionEnabled, err := CheckPgStatMonitorExtensionEnabled(dbConn)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return nil
		}
		if !isPgStatMonitorExtensionEnabled {
			log.Info("Extension 'pg_stat_monitor' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_stat_monitor_extension] = append(extensionDbMap[pg_stat_monitor_extension], dbName)

		isPgWaitSamplingExtensionEnabled, err := CheckPgWaitSamplingExtensionEnabled(dbConn)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return nil
		}
		if !isPgWaitSamplingExtensionEnabled {
			log.Info("Extension 'pg_wait_sampling' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_wait_sampling_extension] = append(extensionDbMap[pg_wait_sampling_extension], dbName)

	}
	return extensionDbMap
}
