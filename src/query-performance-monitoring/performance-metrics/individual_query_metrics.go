package performanceMetrics

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, slowRunningQueries []datamodels.SlowRunningQueryMetrics, pgIntegration *integration.Integration, args args.ArgumentList) []datamodels.IndividualQueryMetrics {
	isExtensionEnabled, err := validations.CheckIndividualQueryMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error validating eligibility for IndividualQueryMetrics: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extensions for PopulateIndividualQueryMetrics is not enabled.")
		return nil
	}
	log.Info("Extensions for PopulateIndividualQueryMetrics is enabled.")
	individualQueryMetricsInterface, individualQueriesForExecPlan := GetIndividualQueryMetrics(conn, args, slowRunningQueries)
	if len(individualQueryMetricsInterface) == 0 {
		log.Info("No individual queries found.")
		return nil
	}
	common_utils.IngestMetric(individualQueryMetricsInterface, "PostgresIndividualQueries", pgIntegration, args)
	return individualQueriesForExecPlan
}

func ConstructIndividualQuery(slowRunningQueries []datamodels.SlowRunningQueryMetrics, args args.ArgumentList) string {
	var queryIDs []string
	for _, query := range slowRunningQueries {
		queryIDs = append(queryIDs, fmt.Sprintf("%d", *query.QueryID))
	}
	query := fmt.Sprintf(queries.IndividualQuerySearch, strings.Join(queryIDs, ","), args.QueryResponseTimeThreshold)
	return query
}

func GetIndividualQueryMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList, slowRunningQueries []datamodels.SlowRunningQueryMetrics) ([]interface{}, []datamodels.IndividualQueryMetrics) {
	query := ConstructIndividualQuery(slowRunningQueries, args)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Info("Error executing query: %v", err)
		return nil, nil
	}
	defer rows.Close()
	anonymizedQueriesByDB := processForAnonymizeQueryMap(slowRunningQueries)
	var individualQueryMetricsForExecPlanList []datamodels.IndividualQueryMetrics
	var individualQueryMetricsListInterface []interface{}
	for rows.Next() {

		var model datamodels.IndividualQueryMetrics
		if err := rows.StructScan(&model); err != nil {
			log.Error("Could not scan row: ", err)
			continue
		}
		individualQueryMetric := model
		anonymizedQueryText := anonymizedQueriesByDB[*model.DatabaseName][*model.QueryID]
		individualQueryMetric.QueryText = &anonymizedQueryText

		model.RealQueryText = model.QueryText
		model.QueryText = &anonymizedQueryText

		individualQueryMetricsForExecPlanList = append(individualQueryMetricsForExecPlanList, model)
		individualQueryMetricsListInterface = append(individualQueryMetricsListInterface, individualQueryMetric)
	}
	return individualQueryMetricsListInterface, individualQueryMetricsForExecPlanList

}

func processForAnonymizeQueryMap(slowRunningQueryMetricsList []datamodels.SlowRunningQueryMetrics) map[string]map[int64]string {
	anonymizeQueryMapByDB := make(map[string]map[int64]string)

	for _, metric := range slowRunningQueryMetricsList {
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
