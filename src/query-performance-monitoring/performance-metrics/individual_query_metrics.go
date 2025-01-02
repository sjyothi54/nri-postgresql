package performancemetrics

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, args args.ArgumentList, databaseNames string) []datamodels.IndividualQueryMetrics {
	isExtensionEnabled, err := validations.CheckIndividualQueryMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Debug("Extension 'pg_stat_monitor' is not enabled.")
		return nil
	}
	log.Info("Extension 'pg_stat_monitor' enabled.")
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, slowRunningQueries, args, databaseNames)
	if len(individualQueryMetricsInterface) == 0 {
		log.Debug("No individual queries found.")
		return nil
	}
	commonutils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, args)
	return individualQueriesForExecPlan
}

func ConstructIndividualQuery(slowRunningQueries []datamodels.SlowRunningQueryMetrics, args args.ArgumentList, databaseNames string) string {
	var queryIDs = make([]string, 0)
	for _, query := range slowRunningQueries {
		queryIDs = append(queryIDs, fmt.Sprintf("%d", *query.QueryID))
	}

	query := fmt.Sprintf(queries.IndividualQuerySearch, strings.Join(queryIDs, ","), databaseNames, args.QueryResponseTimeThreshold)
	return query
}

func GetIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, args args.ArgumentList, databaseNames string) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	if len(slowRunningQueries) == 0 {
		log.Debug("No slow running queries found.")
		return nil, nil
	}
	query := ConstructIndividualQuery(slowRunningQueries, args, databaseNames)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Debug("Error executing query in individual query: %v", err)
		return nil, nil
	}
	anonymizedQueriesByDB := processForAnonymizeQueryMap(slowRunningQueries)
	var individualQueryMetricsForExecPlanList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	for rows.Next() {
		var model datamodels.IndividualQueryMetrics
		if scanErr := rows.StructScan(&model); scanErr != nil {
			log.Error("Could not scan row: ", scanErr)
			continue
		}
		individualQueryMetric := model
		anonymizedQueryText := anonymizedQueriesByDB[*model.DatabaseName][*model.QueryID]
		individualQueryMetric.QueryText = &anonymizedQueryText
		generatedPlanID := commonutils.GenerateRandomIntegerString(*model.QueryID)
		individualQueryMetric.PlanID = generatedPlanID
		model.PlanID = generatedPlanID
		model.RealQueryText = model.QueryText
		model.QueryText = &anonymizedQueryText

		individualQueryMetricsForExecPlanList = append(individualQueryMetricsForExecPlanList, model)
		individualQueryMetricsListInterface = append(individualQueryMetricsListInterface, individualQueryMetric)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, nil
	}
	return individualQueryMetricsListInterface, individualQueryMetricsForExecPlanList
}

func processForAnonymizeQueryMap(queryCPUMetricsList []datamodels.SlowRunningQueryMetrics) map[string]map[int64]string {
	anonymizeQueryMapByDB := make(map[string]map[int64]string)

	for _, metric := range queryCPUMetricsList {
		dbName := *metric.DatabaseName
		queryID := *metric.QueryID
		anonymizedQuery := *metric.QueryText

		if _, exists := anonymizeQueryMapByDB[dbName]; !exists {
			anonymizeQueryMapByDB[dbName] = make(map[int64]string)
		}
		anonymizeQueryMapByDB[dbName][queryID] = anonymizedQuery
	}
	return anonymizeQueryMapByDB
}
