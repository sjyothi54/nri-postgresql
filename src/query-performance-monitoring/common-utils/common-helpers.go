package commonutils

import (
	"fmt"
	"strings"

	"github.com/newrelic/nri-postgresql/src/collection"
)

func GetQuotedStringFromArray(array []string) string {
	var quotedDatabaseNames []string
	for _, name := range array {
		quotedDatabaseNames = append(quotedDatabaseNames, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(quotedDatabaseNames, ",")
}

func GetDatabaseListInString(dbList collection.DatabaseList) string {
	var databaseNames []string
	for dbName := range dbList {
		databaseNames = append(databaseNames, dbName)
	}
	return GetQuotedStringFromArray(databaseNames)
}
