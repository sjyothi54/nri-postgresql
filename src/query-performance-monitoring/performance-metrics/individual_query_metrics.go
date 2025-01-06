package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
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
	log.Debug("Extension 'pg_stat_monitor' enabled.")
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, slowRunningQueries, args, databaseNames)
	if len(individualQueryMetricsInterface) == 0 {
		log.Debug("No individual queries found.")
		return nil
	}
	commonutils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, args)
	return individualQueriesForExecPlan
}

func ConstructIndividualQuery(slowRunningQueries datamodels.SlowRunningQueryMetrics, conn *performancedbconnection.PGSQLConnection, args args.ArgumentList, databaseNames string) string {
	versionSpecificIndividualQuery, _ := commonutils.FetchVersionSpecificIndividualQuieries(conn)
	query := fmt.Sprintf(versionSpecificIndividualQuery, *slowRunningQueries.QueryID, databaseNames, args.QueryResponseTimeThreshold, min(args.QueryCountThreshold, commonutils.MAX_INDIVIDUAL_QUERY_THRESHOLD))
	return query
}

func GetIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, args args.ArgumentList, databaseNames string) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	if len(slowRunningQueries) == 0 {
		log.Debug("No slow running queries found.")
		return nil, nil
	}
	var individualQueryMetricsForExecPlanList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	anonymizedQueriesByDB := processForAnonymizeQueryMap(slowRunningQueries)
	for _, slowRunningMetric := range slowRunningQueries {
		getIndividualQueriesByGroupedQuery(conn, slowRunningMetric, args, databaseNames, anonymizedQueriesByDB, &individualQueryMetricsForExecPlanList, &individualQueryMetricsListInterface)
	}
	return individualQueryMetricsListInterface, individualQueryMetricsForExecPlanList
}

func getIndividualQueriesByGroupedQuery(conn *performancedbconnection.PGSQLConnection, slowRunningQueries datamodels.SlowRunningQueryMetrics, args args.ArgumentList, databaseNames string, anonymizedQueriesByDB map[string]map[int64]string, individualQueryMetricsForExecPlanList *[]datamodels.IndividualQueryMetrics, individualQueryMetricsListInterface *[]interface{}) {

	query := ConstructIndividualQuery(slowRunningQueries, conn, args, databaseNames)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Debug("Error executing query in individual query: %v", err)
		return
	}
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

		*individualQueryMetricsForExecPlanList = append(*individualQueryMetricsForExecPlanList, model)
		*individualQueryMetricsListInterface = append(*individualQueryMetricsListInterface, individualQueryMetric)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return
	}
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
