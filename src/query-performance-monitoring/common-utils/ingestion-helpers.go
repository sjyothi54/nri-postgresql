package commonutils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
)

const publishThreshold = 100
const randomIntRange = 1000000

func SetMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			return
		}
	default:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			return
		}
	}
}

func IngestMetric(metricList []interface{}, eventName string, pgIntegration *integration.Integration, args args.ArgumentList) {
	metricCount := 0
	lenOfMetricList := len(metricList)
	instanceEntity, entityError := pgIntegration.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
	if entityError != nil {
		log.Error("Error creating instance entity: %v", entityError)
		return
	}
	for _, model := range metricList {
		if model == nil {
			continue
		}
		metricCount += 1
		metricSet := instanceEntity.NewMetricSet(eventName)

		modelValue := reflect.ValueOf(model)
		if modelValue.Kind() == reflect.Ptr {
			modelValue = modelValue.Elem()
		}
		if !modelValue.IsValid() || modelValue.Kind() != reflect.Struct {
			continue
		}
		modelType := reflect.TypeOf(model)
		setMetricParser(modelValue, modelType, metricSet)

		if metricCount == publishThreshold || metricCount == lenOfMetricList {
			metricCount = 0
			publishErr := pgIntegration.Publish()
			if publishErr != nil {
				log.Error("Error publishing metrics: %v", entityError)
				return
			}
			instanceEntity, entityError = pgIntegration.Entity(fmt.Sprintf("%s:%s", "localhost", "5432"), "pg-instance")
			if entityError != nil {
				log.Error("Error creating instance entity: %v", entityError)
				return
			}
		}
	}
	if metricCount > 0 {
		publishError := pgIntegration.Publish()
		if publishError != nil {
			log.Error("Error publishing metrics: %v", entityError)
			return
		}
	}
}

func setMetricParser(modelValue reflect.Value, modelType reflect.Type, metricSet *metric.Set) {
	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")
		ingestData := fieldType.Tag.Get("ingest_data")

		if ingestData == "false" {
			log.Info("not ingesting data for field: %s", metricName)
			continue
		}

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			SetMetric(metricSet, metricName, field.Interface(), sourceType)
		}
	}
}

func GenerateRandomIntegerString(queryID int64) *string {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(randomIntRange))
	if err != nil {
		return nil
	}
	currentTime := time.Now().Format("20060102150405")
	result := fmt.Sprintf("%d-%d-%s", queryID, randomInt.Int64(), currentTime)
	return &result
}
