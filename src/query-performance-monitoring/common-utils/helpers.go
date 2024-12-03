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
		metricSet.SetMetric(name, value, metric.GAUGE)
	case `attribute`:
		metricSet.SetMetric(name, value, metric.ATTRIBUTE)
	default:
		metricSet.SetMetric(name, value, metric.GAUGE)
	}
}

func FatalIfError(err error) {
	if err != nil {
		fmt.Errorf("Error: %v", err)
		panic(err)
	}
}

func SetMetricsParser(instanceEntity *integration.Entity, eventName string, args args.ArgumentList, model interface{}) {
	metricSet := CreateMetricSet(instanceEntity, eventName, args)
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)
	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			fmt.Print("Field is a pointer")
			SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			fmt.Println("Field is not a pointer")
			SetMetric(metricSet, metricName, field.Interface(), sourceType)
		}
	}
}
