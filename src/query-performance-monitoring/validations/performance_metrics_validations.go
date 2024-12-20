package validations

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
)

const pg_stat_statements_extension = "pg_stat_statements"
const pg_stat_monitor_extension = "pg_stat_monitor"
const pg_wait_sampling_extension = "pg_wait_sampling"

var extensionDbMap = make(map[string][]string)

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

func CheckDbWithWaitMetricsEligibility() ([]*performanceDbConnection.PGSQLConnection, error) {
	dbWithPgWaitExtension := extensionDbMap[pg_wait_sampling_extension]
	dbWIthPgStatExtension := extensionDbMap[pg_stat_statements_extension]
	dbStatMap := make(map[string]bool)
	for _, db := range dbWIthPgStatExtension {
		dbStatMap[db] = true
	}
	var commonDbs []string
	for _, db := range dbWithPgWaitExtension {
		if dbStatMap[db] {
			commonDbs = append(commonDbs, db)
		}
	}
	var dbConnections []*performanceDbConnection.PGSQLConnection
	for _, dbName := range commonDbs {
		dbConnections = append(dbConnections, performanceDbConnection.DbConnections[dbName])
	}
	return dbConnections, nil

}

func CheckDbsWithSlowQueryMetricsEligibility() ([]*performanceDbConnection.PGSQLConnection, error) {
	dbWithPgStatExtension := extensionDbMap[pg_stat_statements_extension]
	var dbConnections []*performanceDbConnection.PGSQLConnection
	for _, dbName := range dbWithPgStatExtension {
		dbConnections = append(dbConnections, performanceDbConnection.DbConnections[dbName])
	}
	return dbConnections, nil
}

func CheckDbsWithBlockingSessionMetricsEligibility() ([]*performanceDbConnection.PGSQLConnection, error) {
	return CheckDbsWithSlowQueryMetricsEligibility()
}

func CheckDbsWithIndividualQueryMetricsEligibility() ([]*performanceDbConnection.PGSQLConnection, error) {
	dbWithPgStatMonitorExtension := extensionDbMap[pg_stat_monitor_extension]
	var dbConnections []*performanceDbConnection.PGSQLConnection
	for _, dbName := range dbWithPgStatMonitorExtension {
		dbConnections = append(dbConnections, performanceDbConnection.DbConnections[dbName])
	}
	return dbConnections, nil
}

func GetExtensionEnabledDbList() {

	extensionDbMap = make(map[string][]string)
	for dbName, dbConn := range performanceDbConnection.DbConnections {

		isPgStatStatementExtensionEnabled, err := checkIsExtensionEnabled(dbConn, pg_stat_statements_extension)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return
		}
		if !isPgStatStatementExtensionEnabled {
			log.Info("Extension 'pg_stat_statements' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_stat_statements_extension] = append(extensionDbMap[pg_stat_statements_extension], dbName)

		isPgStatMonitorExtensionEnabled, err := checkIsExtensionEnabled(dbConn, pg_stat_monitor_extension)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return
		}
		if !isPgStatMonitorExtensionEnabled {
			log.Info("Extension 'pg_stat_monitor' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_stat_monitor_extension] = append(extensionDbMap[pg_stat_monitor_extension], dbName)

		isPgWaitSamplingExtensionEnabled, err := checkIsExtensionEnabled(dbConn, pg_wait_sampling_extension)
		if err != nil {
			log.Error("Error executing query: %v", err)
			return
		}
		if !isPgWaitSamplingExtensionEnabled {
			log.Info("Extension 'pg_wait_sampling' not enabled for %s.", dbName)
			continue
		}
		extensionDbMap[pg_wait_sampling_extension] = append(extensionDbMap[pg_wait_sampling_extension], dbName)

	}
}
