package query_metrics

import (
	"errors"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func getSlowRunningMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.SlowRunningQuery, []*int64, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var queryIdList []*int64
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
		queryIdList = append(queryIdList, slowQuery.QueryID)

	}

	fmt.Print("Slow Queries: ", slowQueries)

	return slowQueries, queryIdList, nil
}

func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) ([]datamodels.SlowRunningQuery, []*int64, error) {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil, nil, err
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return nil, nil, errors.New("extension 'pg_stat_statements' is not enabled")
	}
	log.Info("Extension 'pg_stat_statements' enabled.")
	slowQueries, queryIdList, err := getSlowRunningMetrics(conn)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil, nil, err
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil, nil, errors.New("no slow-running queries found")
	}

	for _, model := range slowQueries {
		common_utils.SetMetricsParser(instanceEntity, "PostgresSlowQueriesV18", args, model)
	}

	return slowQueries, queryIdList, nil
}
