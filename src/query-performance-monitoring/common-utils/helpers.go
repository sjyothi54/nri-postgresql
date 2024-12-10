package common_utils

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"reflect"
)

func CreateMetricSet(e *integration.Entity, sampleName string, args args.ArgumentList) *metric.Set {
	return metricSet(
		e,
		sampleName,
		args.Hostname,
		args.Port,
	)
}

func metricSet(e *integration.Entity, eventType, hostname string, port string) *metric.Set {
	return e.NewMetricSet(
		eventType,
		attribute.Attr("hostname", hostname),
		attribute.Attr("port", port),
	)
}

func SetMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			fmt.Println("Error in setting metric", err)
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			fmt.Println("Error in setting metric", err)
			return
		}
	default:
		fmt.Println("Error: metric type not supported")
		return

	}
}

func SetMetricsParser(instanceEntity *integration.Entity, eventName string, args args.ArgumentList, pgIntegration *integration.Integration, metricList []interface{}) {
	fmt.Println("Started to ingest dataa")
	//pgIntegrationNew, err := integration.New("com.newrelic.postgresql", "0.0.0", integration.Args(&args))
	//fmt.Println("pgIntegrationNew")
	//instance, err := pgIntegrationNew.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
	//
	//if err != nil {
	//	fmt.Println("Error in creating integration", err)
	//	return
	//}

	lenOfMetric := len(metricList)
	cnt := 0
	for _, model := range metricList {
		metricSetIngestion := CreateMetricSet(instanceEntity, eventName, args)

		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)

		for i := 0; i < modelValue.NumField(); i++ {
			cnt += 1
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				SetMetric(metricSetIngestion, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				SetMetric(metricSetIngestion, metricName, field.Interface(), sourceType)
			}
			//fmt.Println("byee", cnt)
			if cnt == 60 || cnt == lenOfMetric {
				//fmt.Println("heyyyy", lenOfMetric, cnt, metricSetIngestion.Metrics)
				fmt.Println("[SetMetricsParser] Before Publish Entities", len(pgIntegration.Entities))
				err := pgIntegration.Publish()
				if err != nil {
					fmt.Println("Error in publishing metrics", err)
					return
				}
				cnt = 0
				//fmt.Println("entitiesssss", pgIntegration.Entities)
				fmt.Println("[SetMetricsParser] After Publish Entities", len(pgIntegration.Entities))
				pgIntegration.Entities = append(pgIntegration.Entities, instanceEntity)
			}
		}
	}
}
