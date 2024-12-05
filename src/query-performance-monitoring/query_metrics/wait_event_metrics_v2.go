// src/query-performance-monitoring/query_metrics/populate_wait_event_metrics_v2.go
package query_metrics

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"

	"github.com/newrelic/nri-postgresql/src/args"
)

// PopulateWaitEventMetricsV2 collects wait event metrics and adds them to the New Relic integration entity.
func PopulateWaitEventMetricsV2(instanceEntity *integration.Entity, db *sqlx.DB, cmdArgs args.ArgumentList) error {
	// Define the wait events query
	query := `
    SELECT 
        wait_event AS wait_event_name,
        wait_event_type AS wait_category,
        total_wait_time AS total_wait_time_ms,
        waiting_tasks AS waiting_tasks_count,
        now() AS collection_timestamp,
        queryid AS query_id,
        query AS query_text,
        datname AS database_name
    FROM 
        pg_wait_sampling_profile;
    `

	// Execute the query
	results, err := ExecuteQuery(db, query)
	if err != nil {
		log.Error("Error executing wait event query: %v", err)
		return fmt.Errorf("failed to execute wait event query: %w", err)
	}

	if len(results) == 0 {
		log.Info("No wait event metrics found.")
		return nil
	}

	log.Info("Found %d wait event metrics.", len(results))

	// Iterate over each row and add metrics to the New Relic metric sets
	for _, row := range results {
		// Create a new metric set for each wait event sample
		metricSet := instanceEntity.NewMetricSet("PostgresqlWaitEventSample")

		// Set the mandatory event_type attribute
		if err := metricSet.SetMetric("event_type", "PostgresqlWaitEventSample", metric.ATTRIBUTE); err != nil {
			log.Error("Error setting event_type attribute: %v", err)
			return fmt.Errorf("failed to set event_type: %w", err)
		}

		// Iterate over the row map and set metrics
		for key, value := range row {
			// Skip the event_type as it's already set
			if key == "event_type" {
				continue
			}

			// Determine the metric type based on the value's type
			metricType := determineMetricType(value)

			// Set the metric
			if err := metricSet.SetMetric(key, value, metricType); err != nil {
				log.Warn("Failed to set metric '%s': %v", key, err)
				// Continue setting other metrics even if one fails
			}
		}
	}

	return nil
}

// ExecuteQuery executes a SQL query and returns the results as a slice of maps.
func ExecuteQuery(db *sqlx.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Warn("Error closing rows: %v", cerr)
		}
	}()

	var results []map[string]interface{}

	for rows.Next() {
		rowData := make(map[string]interface{})
		if err := rows.MapScan(rowData); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		// Convert SQL types to native Go types
		for key, value := range rowData {
			rowData[key] = ConvertSQLTypes(value)
		}

		results = append(results, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// ConvertSQLTypes converts SQL types to native Go types, handling NULLs gracefully.
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
		log.Warn("Unexpected type %T for value %v", v, v)
		return nil
	}
}

// determineMetricType determines the New Relic metric type based on the Go type of the value.
func determineMetricType(value interface{}) metric.SourceType {
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return metric.GAUGE
	case string, bool, time.Time:
		return metric.ATTRIBUTE
	default:
		return metric.ATTRIBUTE
	}
}
