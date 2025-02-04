package validations

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
)

var extensions map[string]bool

func fetchAllExtensions(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) error {
	rows, err := conn.Queryx("SELECT extname FROM pg_extension", app)
	if err != nil {
		log.Error("Error executing query: ", err.Error())
		return err
	}
	defer rows.Close()
	extensions = make(map[string]bool)
	for rows.Next() {
		var extname string
		if err := rows.Scan(&extname); err != nil {
			log.Error("Error scanning rows: ", err.Error())
			return err
		}
		extensions[extname] = true
	}
	return nil
}

func isExtensionEnabled(extensionName string) bool {
	return extensions[extensionName]
}

func CheckSlowQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	loadExtensionsMap(conn, app)
	return isExtensionEnabled("pg_stat_statements"), nil
}

func CheckWaitEventMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	loadExtensionsMap(conn, app)
	return isExtensionEnabled("pg_wait_sampling") && isExtensionEnabled("pg_stat_statements"), nil
}

func CheckBlockingSessionMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, version uint64, app *newrelic.Application) (bool, error) {
	// Version 12 and 13 do not require the pg_stat_statements extension
	if version == commonutils.PostgresVersion12 || version == commonutils.PostgresVersion13 {
		return true, nil
	}
	loadExtensionsMap(conn, app)
	return isExtensionEnabled("pg_stat_statements"), nil
}

func CheckIndividualQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	loadExtensionsMap(conn, app)
	return isExtensionEnabled("pg_stat_monitor"), nil
}

func CheckPostgresVersionSupportForQueryMonitoring(version uint64) bool {
	return version >= commonutils.PostgresVersion12
}

func ClearExtensionsLoadCache() {
	extensions = nil
}

func loadExtensionsMap(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) {
	if extensions == nil {
		if err := fetchAllExtensions(conn, app); err != nil {
			log.Error("Error fetching all extensions: %v", err)
		}
	}
}
