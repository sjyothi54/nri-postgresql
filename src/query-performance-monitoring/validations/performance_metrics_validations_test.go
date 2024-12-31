package validations

import (
	"testing"

	"fmt"

	performancedbconnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_isExtensionEnabled(t *testing.T) {
	conn, mock := performancedbconnection.CreateMockSQL(t)

	tests := []struct {
		name          string
		extensionName string
		mockRows      *sqlmock.Rows
		mockError     error
		expected      bool
		expectError   bool
	}{
		{
			name:          "Extension enabled",
			extensionName: "pg_stat_statements",
			mockRows:      sqlmock.NewRows([]string{"count"}).AddRow(1),
			expected:      true,
			expectError:   false,
		},
		{
			name:          "Extension not enabled",
			extensionName: "pg_stat_statements",
			mockRows:      sqlmock.NewRows([]string{"count"}).AddRow(0),
			expected:      false,
			expectError:   false,
		},
		{
			name:          "Query error",
			extensionName: "pg_stat_statements",
			mockError:     fmt.Errorf("query error"),
			expected:      false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockError != nil {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = .*").WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = .*").WillReturnRows(tt.mockRows)
			}

			result, err := isExtensionEnabled(conn, tt.extensionName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_CheckPgWaitSamplingExtensionEnabled(t *testing.T) {
	conn, mock := performancedbconnection.CreateMockSQL(t)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_wait_sampling'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := CheckPgWaitSamplingExtensionEnabled(conn)
	assert.NoError(t, err)
	assert.True(t, result)
}

func Test_CheckPgStatStatementsExtensionEnabled(t *testing.T) {
	conn, mock := performancedbconnection.CreateMockSQL(t)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_statements'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := CheckPgStatStatementsExtensionEnabled(conn)
	assert.NoError(t, err)
	assert.True(t, result)
}

func Test_CheckPgStatMonitorExtensionEnabled(t *testing.T) {
	conn, mock := performancedbconnection.CreateMockSQL(t)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM pg_extension WHERE extname = 'pg_stat_monitor'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := CheckPgStatMonitorExtensionEnabled(conn)
	assert.NoError(t, err)
	assert.True(t, result)
}
