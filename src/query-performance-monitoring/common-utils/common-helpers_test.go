
package commonutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/newrelic/nri-postgresql/src/collection"
)

func TestGetQuotedStringFromArray(t *testing.T) {
	array := []string{"db1", "db2", "db3"}
	expected := "'db1','db2','db3'"
	result := getQuotedStringFromArray(array)
	assert.Equal(t, expected, result)
}

func TestGetDatabaseListInString(t *testing.T) {
	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{},
		"db2": collection.SchemaList{},
	}
	expected := "'db1','db2'"
	result := GetDatabaseListInString(dbList)
	assert.Equal(t, expected, result)
}

func TestAnonymizeQueryText(t *testing.T) {
	query := "SELECT * FROM table WHERE id = 123 AND name = 'John'"
	expected := "SELECT * FROM table WHERE id = ? AND name = ?"
	result := AnonymizeQueryText(query)
	assert.Equal(t, expected, result)
}

func TestGeneratePlanID(t *testing.T) {
	queryID := "query123"
	result := GeneratePlanID(queryID)
	assert.NotNil(t, result)
	assert.Contains(t, *result, queryID)
	assert.Contains(t, *result, time.Now().Format(TimeFormat))
}