package validations_test

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCheckBlockingSessionMetricsFetchEligibilityExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(12)
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_monitor"))
	isExtensionEnabledTest, _ := validations.CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	isExtensionEnabledTest, _ := validations.CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestCheckWaitEventMetricsFetchEligibility(t *testing.T) {
	version := uint64(15)
	validationQuery := fmt.Sprintf("SELECT extname FROM pg_extension")
	testCases := []struct {
		waitExt  string
		statExt  string
		expected bool
	}{
		{"pg_wait_sampling", "pg_stat_statements", true}, // Success
		{"pg_wait_sampling", "", false},                  // Fail V1
		{"", "pg_stat_statements", false},                // Fail V2
	}

	conn, mock := connection.CreateMockSQL(t)
	for _, tc := range testCases {
		mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow(tc.waitExt).AddRow(tc.statExt))
		isExtensionEnabledTest, _ := validations.CheckWaitEventMetricsFetchEligibility(conn, version)
		assert.Equal(t, isExtensionEnabledTest, tc.expected)
		assert.NoError(t, mock.ExpectationsWereMet())
		validations.ClearExtensionsLoadCache()
	}
}

func TestCheckSlowQueryMetricsFetchEligibilityUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := validations.CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	isExtensionEnabledTest, _ := validations.CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT extname FROM pg_extension")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	isExtensionEnabledTest, _ := validations.CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
	validations.ClearExtensionsLoadCache()
}
