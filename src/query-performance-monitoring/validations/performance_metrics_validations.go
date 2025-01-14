package validations

import (
	"fmt"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
)

func isExtensionEnabled(conn *performancedbconnection.PGSQLConnection, extensionName string, app *newrelic.Application) (bool, error) {
	rows, err := conn.Queryx(fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", extensionName), app)
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

func CheckSlowQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_statements", app)
}

func CheckWaitEventMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	pgWaitExtension, waitErr := isExtensionEnabled(conn, "pg_wait_sampling", app)
	if waitErr != nil {
		return false, waitErr
	}
	pgStatExtension, statErr := isExtensionEnabled(conn, "pg_stat_statements", app)
	if statErr != nil {
		return false, statErr
	}
	return pgWaitExtension && pgStatExtension, nil
}

func CheckBlockingSessionMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, version uint64, app *newrelic.Application) (bool, error) {
	if version == commonutils.PostgresVersion12 || version == commonutils.PostgresVersion13 {
		return true, nil
	}
	return isExtensionEnabled(conn, "pg_stat_statements", app)
}

func CheckIndividualQueryMetricsFetchEligibility(conn *performancedbconnection.PGSQLConnection, app *newrelic.Application) (bool, error) {
	return isExtensionEnabled(conn, "pg_stat_monitor", app)
}
