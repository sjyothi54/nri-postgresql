package performance_metrics

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, pgIntegration *integration.Integration, args args.ArgumentList) {

	if len(results) == 0 {
		log.Info("No individual queries found.")
		return
	}

	executionDetailsList := GetExecutionPlanMetrics(results, args)

	common_utils.IngestMetric(executionDetailsList, "PostgresExecutionPlanMetrics", pgIntegration, args)
}

func GetExecutionPlanMetrics(results []datamodels.IndividualQueryMetrics, args args.ArgumentList) []interface{} {

	var executionPlanMetricsList []interface{}

	processExecutionPlanOfQueries(results, &executionPlanMetricsList)

	return executionPlanMetricsList

}

func processExecutionPlanOfQueries(individualQueriesList []datamodels.IndividualQueryMetrics, executionPlanMetricsList *[]interface{}) {

	for _, individualQuery := range individualQueriesList {
		query := "EXPLAIN (FORMAT JSON) " + *individualQuery.RealQueryText
		log.Info("Execution Plan Query : %s", query)
		rows, err := performanceDbConnection.DbConnections[*individualQuery.DatabaseName].Queryx(query)
		if err != nil {
			log.Info("Error executing query: %v", err)
			continue
		}
		defer rows.Close()
		if !rows.Next() {
			log.Info("Execution plan not found for queryId", *individualQuery.QueryId)
			continue
		}
		var execPlanJSON string
		if err := rows.Scan(&execPlanJSON); err != nil {
			log.Error("Error scanning row: ", err.Error())
			continue
		}

		var execPlan []map[string]interface{}
		err = json.Unmarshal([]byte(execPlanJSON), &execPlan)
		if err != nil {
			log.Error("Failed to unmarshal execution plan: %v", err)
			continue
		}
		fetchNestedExecutionPlanDetails(individualQuery, 0, execPlan[0]["Plan"].(map[string]interface{}), executionPlanMetricsList)
	}
}

func fetchNestedExecutionPlanDetails(individualQuery datamodels.IndividualQueryMetrics, level int, execPlan map[string]interface{}, executionPlanMetricsList *[]interface{}) {
	var execPlanMetrics datamodels.QueryExecutionPlanMetrics
	err := mapstructure.Decode(execPlan, &execPlanMetrics)
	if err != nil {
		log.Error("Failed to decode execPlan to execPlanMetrics: %v", err)
		return
	}
	execPlanMetrics.QueryText = *individualQuery.QueryText
	execPlanMetrics.QueryId = *individualQuery.QueryId
	execPlanMetrics.DatabaseName = *individualQuery.DatabaseName
	execPlanMetrics.Level = level
	if individualQuery.PlanId != nil {
		execPlanMetrics.PlanId = *individualQuery.PlanId
	} else {
		execPlanMetrics.PlanId = 999
	}

	*executionPlanMetricsList = append(*executionPlanMetricsList, execPlanMetrics)

	if nestedPlans, ok := execPlan["Plans"].([]interface{}); ok {
		for _, nestedPlan := range nestedPlans {
			if nestedPlanMap, ok := nestedPlan.(map[string]interface{}); ok {
				fetchNestedExecutionPlanDetails(individualQuery, level+1, nestedPlanMap, executionPlanMetricsList)
			}
		}
	}
}
