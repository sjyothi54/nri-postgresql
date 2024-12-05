package query_metrics

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"reflect"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performance_db_connection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-db-connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

// ConvertSQLTypes converts SQL driver types to native Go types, handling NULL values gracefully.
func ConvertSQLTypes(value interface{}) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case bool, int64, float64, string:
		return v
	case []byte:
		return string(v)
	case time.Time:
		return v
	case sql.NullBool:
		if v.Valid {
			return v.Bool
		}
		return nil
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}
		return nil
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return nil
	case sql.NullTime:
		if v.Valid {
			return v.Time
		}
		return nil
	default:
		// Handle other sql.Null* types using reflection
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Struct {
			validField := val.FieldByName("Valid")
			if validField.IsValid() && validField.Bool() {
				dataField := val.Field(0)
				return dataField.Interface()
			}
		}
		return v
	}
}

// ExecuteQuery executes a SQL query and returns the results as a slice of map[string]interface{}.
// It handles type mismatches gracefully and includes detailed error handling.
func ExecuteQuery(conn *performance_db_connection.PGSQLConnection, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Error("Error closing rows: %v", closeErr)
		}
	}()

	var results []map[string]interface{}

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns: %w", err)
		}

		scanArgs := make([]interface{}, len(columns))
		rawValues := make([]interface{}, len(columns))

		for i := range scanArgs {
			scanArgs[i] = &rawValues[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			rawValue := rawValues[i]
			rowData[col] = ConvertSQLTypes(rawValue)
		}

		results = append(results, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// SetMetricsParser processes metrics and adds them to the integration entity.
func SetMetricsParser(entity *integration.Entity, metricSetName string, args args.ArgumentList, data map[string]interface{}) error {
	metricSet := entity.NewMetricSet(metricSetName)
	for key, value := range data {
		if err := metricSet.SetMetric(key, value, getMetricType(value)); err != nil {
			return fmt.Errorf("failed to set metric %s: %w", key, err)
		}
	}
	return nil
}

func getMetricType(value interface{}) metric.SourceType {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return metric.GAUGE
	case uint, uint8, uint16, uint32, uint64:
		return metric.GAUGE
	case float32, float64:
		return metric.GAUGE
	case bool:
		return metric.ATTRIBUTE
	case string:
		return metric.ATTRIBUTE
	case time.Time:
		return metric.ATTRIBUTE
	default:
		return metric.ATTRIBUTE
	}
}

// PopulateWaitEventMetrics collects wait event metrics and populates them into the integration entity.
// It uses a generic query execution approach and includes comprehensive error handling and logging.
func PopulateWaitEventMetricsV2(instanceEntity *integration.Entity, conn *performance_db_connection.PGSQLConnection, args args.ArgumentList) error {
	isExtensionEnabled, err := validations.CheckPgWaitExtensionEnabled(conn)
	if err != nil {
		log.Error("Error checking 'pg_wait_sampling' extension: %v", err)
		return err
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_wait_sampling' is not enabled.")
		return errors.New("extension 'pg_wait_sampling' is not enabled")
	}
	log.Info("Extension 'pg_wait_sampling' is enabled.")

	// Execute the wait event query using the generic ExecuteQuery function
	query := queries.WaitEvents
	waitEventMetrics, err := ExecuteQuery(conn, query)
	if err != nil {
		log.Error("Error executing wait event query: %v", err)
		return err
	}

	if len(waitEventMetrics) == 0 {
		log.Info("No wait event metrics found.")
		return nil
	}

	log.Debug("WaitEventMetrics: %+v", waitEventMetrics)

	for _, metric := range waitEventMetrics {
		// Process each metric using the SetMetricsParser function
		err := SetMetricsParser(instanceEntity, "PostgresqlWaitEventMetricsV1", args, metric)
		if err != nil {
			log.Error("Error setting metrics parser for metric %v: %v", metric, err)
			// Continue processing other metrics
			continue
		}
	}

	return nil
}
