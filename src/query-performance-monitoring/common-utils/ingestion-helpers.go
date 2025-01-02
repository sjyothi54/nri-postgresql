package commonutils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
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
	instanceEntity, err := createEntity(pgIntegration, args)
	if err != nil {
		log.Error("Error creating entity: %v", err)
		return
	}

	metricCount := 0
	lenOfMetricList := len(metricList)

	for _, model := range metricList {
		if model == nil {
			continue
		}
		metricCount += 1
		metricSet := instanceEntity.NewMetricSet(eventName)

		processModel(model, metricSet)

		if metricCount == publishThreshold || metricCount == lenOfMetricList {
			metricCount = 0
			if err := publishMetrics(pgIntegration, &instanceEntity, args); err != nil {
				log.Error("Error publishing metrics: %v", err)
				return
			}
		}
	}

	if err := publishMetrics(pgIntegration, &instanceEntity, args); err != nil {
		log.Error("Error publishing metrics: %v", err)
		return
	}
}

func createEntity(pgIntegration *integration.Integration, args args.ArgumentList) (*integration.Entity, error) {
	return pgIntegration.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
}

func processModel(model interface{}, metricSet *metric.Set) {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}
	if !modelValue.IsValid() || modelValue.Kind() != reflect.Struct {
		return
	}

	modelType := reflect.TypeOf(model)

	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")
		ingestData := fieldType.Tag.Get("ingest_data")

		if ingestData == "false" {
			continue
		}

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			SetMetric(metricSet, metricName, field.Interface(), sourceType)
		}
	}
}

func publishMetrics(pgIntegration *integration.Integration, instanceEntity **integration.Entity, args args.ArgumentList) error {
	if err := pgIntegration.Publish(); err != nil {
		return err
	}
	var err error
	*instanceEntity, err = pgIntegration.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
	return err
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

func FetchVersion(conn *performancedbconnection.PGSQLConnection) (int, error) {
	var versionStr string
	rows, err := conn.Queryx("SELECT version()")
	if err != nil {
		log.Error("Error executing query: %v", err)
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&versionStr); err != nil {
			log.Error("Error scanning version: %v", err)
			return 0, err
		}
	}

	log.Info("versionstring", versionStr)

	re := regexp.MustCompile(`PostgreSQL (\d+)\.`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 2 {
		log.Error("Unable to parse PostgreSQL version from string: %s", versionStr)
		return 0, fmt.Errorf("unable to parse PostgreSQL version from string: %s", versionStr)
	}

	version, err := strconv.Atoi(matches[1])
	log.Info("version", version)
	if err != nil {
		log.Error("Error converting version to integer: %v", err)
		return 0, err
	}
	return version, nil
	// switch {
	// case version == 12:
	// 	return queries.SlowQueriesForV12, nil
	// case version >= 13:
	// 	return queries.SlowQueriesForV13AndAbove, nil
	// default:
	// 	return "", fmt.Errorf("unsupported PostgreSQL version: %d", version)
	// }
}

func FetchVersionSpecificSlowQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == 12:
		return queries.SlowQueriesForV12, nil
	case version >= 13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", fmt.Errorf("unsupported PostgreSQL version: %d", version)
	}
}
