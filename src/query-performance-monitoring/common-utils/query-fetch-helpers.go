package commonutils

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func FetchVersion(conn *performancedbconnection.PGSQLConnection) (int, error) {
	var versionStr string
	rows, err := conn.Queryx("SELECT version()")
	if err != nil {
		log.Error("Error executing query: %v", err)
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.New("no rows returned from version query")
	}
	if err := rows.Scan(&versionStr); err != nil {
		log.Error("Error scanning version: %v", err)
		return 0, err
	}
	re := regexp.MustCompile(VERSION_REGEX)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < VERSION_INDEX {
		log.Error("Unable to parse PostgreSQL version from string: %s", versionStr)
		return 0, PARSE_VERSION_ERROR
	}

	version, err := strconv.Atoi(matches[1])
	log.Debug("version", version)
	if err != nil {
		log.Error("Error converting version to integer: %v", err)
		return 0, err
	}
	return version, nil
}

func FetchVersionSpecificSlowQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == POSTGRES_VERSION_12:
		return queries.SlowQueriesForV12, nil
	case version >= POSTGRES_VERSION_13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", UNSUPPORTED_VERSION
	}
}

func FetchVersionSpecificBlockingQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == POSTGRES_VERSION_12, version == POSTGRES_VERSION_13:
		return queries.BlockingQueriesForV12AndV13, nil
	case version >= POSTGRES_VERSION_14:
		return queries.BlockingQueriesForV14AndAbove, nil
	default:
		return "", UNSUPPORTED_VERSION
	}
}

func FetchVersionSpecificIndividualQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == POSTGRES_VERSION_12:
		return queries.IndividualQuerySearchV12, nil
	case version >= POSTGRES_VERSION_12:
		return queries.IndividualQuerySearchV13AndAbove, nil
	default:
		return "", UNSUPPORTED_VERSION
	}
}
