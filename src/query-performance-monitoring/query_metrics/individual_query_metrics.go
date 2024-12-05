package query_metrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"strings"
)

func getIndividualMetrics(conn *performance_db_connection.PGSQLConnection) ([]datamodels.QueryPlanMetrics, error) {
	var individualQueryMetricList []datamodels.QueryPlanMetrics
	var individualQuerySearch = queries.IndividualQuerySearch
	individualQueriesRows, err := conn.Queryx(individualQuerySearch)
	if err != nil {
		fmt.Printf("Error in fetching individual query metrics: %v", err)
		return nil, err
	}
	for individualQueriesRows.Next() {
		var individualQueryMetric datamodels.QueryPlanMetrics
		if err := individualQueriesRows.StructScan(&individualQueryMetric); err != nil {
			fmt.Printf("Failed to scan query metrics row: %v\n", err)
			return nil, err
		}
		individualQueryMetricList = append(individualQueryMetricList, individualQueryMetric)
	}
	return individualQueryMetricList, nil
}

func PopulateIndividualMetrics(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList, queryIDList []*int64) ([]datamodels.QueryPlanMetrics, error) {
	if len(queryIDList) == 0 {
		log.Warn("queryIDList is empty")
		return nil, nil
	}
	query := "SELECT queryId, query FROM pg_stat_monitor WHERE query like 'select * from actor%' and queryId IN ("

	var idStrings []string
	for _, id := range queryIDList {
		if id != nil {
			idStrings = append(idStrings, fmt.Sprintf("%d", *id))
		}
	}

	// Finalize the query string
	query += strings.Join(idStrings, ", ") + ")"

	fmt.Println("query: ", query)

	test := common_utils.CreateMetricSet(instanceEntity, "PostgresIndividualQueriesV19", args)
	err := test.SetMetric("test", "test", metric.ATTRIBUTE)

	individualQueriesMetricsList, err := getIndividualMetrics(conn)
	if err != nil {
		return nil, err
	}

	test1 := common_utils.CreateMetricSet(instanceEntity, "PostgresIndividualQueriesV18", args)
	err = test1.SetMetric("test", "test", metric.ATTRIBUTE)
	if err != nil {
		return nil, err
	}

	var queryTextRow1 = individualQueriesMetricsList[0].QueryText
	fmt.Print("queryTextRow1: ", queryTextRow1)

	test3 := common_utils.CreateMetricSet(instanceEntity, "PostgresIndividualQueriesV22", args)
	err = test3.SetMetric("queryText", "teeeee", metric.ATTRIBUTE)
	if err != nil {
		return nil, err
	}

	for _, model := range individualQueriesMetricsList {
		fmt.Println("model", model.QueryText)
		common_utils.SetMetricsParser(instanceEntity, "PostgresqlIndividualMetricsV1", args, model)

		//metricSetIngestion := instanceEntity.NewMetricSet("PostgresIndividualQueriesV18")
		//modelValue := reflect.ValueOf(model)
		//modelType := reflect.TypeOf(model)
		//for i := 0; i < modelValue.NumField(); i++ {
		//	field := modelValue.Field(i)
		//	fieldType := modelType.Field(i)
		//	metricName := fieldType.Tag.Get("metric_name")
		//	sourceType := fieldType.Tag.Get("source_type")
		//
		//	if field.Kind() == reflect.Ptr && !field.IsNil() {
		//		common_utils.SetMetric(metricSetIngestion, metricName, field.Elem().Interface(), sourceType)
		//	} else if field.Kind() != reflect.Ptr {
		//		common_utils.SetMetric(metricSetIngestion, metricName, field.Interface(), sourceType)
		//	}
		//}
		break
	}

	return individualQueriesMetricsList, nil
}
