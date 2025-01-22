package validations

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
)

func TestCheckBlockingSessionMetricsFetchEligibilityExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(12)
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilityUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckBlockingSessionMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilityUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_monitor")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_monitor")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	isExtensionEnabledTest, _ := CheckIndividualQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibilityUnsupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(15)
	validationWait := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	mock.ExpectQuery(regexp.QuoteMeta(validationWait)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	validationStat := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationStat)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibilitySupportedVersionFailV1(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(15)
	validationWait := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	mock.ExpectQuery(regexp.QuoteMeta(validationWait)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	validationStat := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationStat)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibilitySupportedVersionFailV2(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(15)
	validationWait := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	mock.ExpectQuery(regexp.QuoteMeta(validationWait)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	validationStat := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationStat)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	isExtensionEnabledTest, _ := CheckWaitEventMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilityUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(11)
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	isExtensionEnabledTest, _ := CheckSlowQueryMetricsFetchEligibility(conn, version)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}
