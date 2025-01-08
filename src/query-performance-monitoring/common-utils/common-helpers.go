package commonutils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newrelic/nri-postgresql/src/collection"
)

func GetQuotedStringFromArray(array []string) string {
	var quotedDatabaseNames = make([]string, 0)
	for _, name := range array {
		quotedDatabaseNames = append(quotedDatabaseNames, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(quotedDatabaseNames, ",")
}

func GetDatabaseListInString(dbList collection.DatabaseList) string {
	var databaseNames = make([]string, 0)
	for dbName := range dbList {
		databaseNames = append(databaseNames, dbName)
	}
	if len(databaseNames) == 0 {
		return ""
	}
	return GetQuotedStringFromArray(databaseNames)
}

func AnonymizeQueryText(query string) string {
	re := regexp.MustCompile(`'[^']*'|\d+|".*?"`)
	anonymizedQuery := re.ReplaceAllString(query, "?")
	return anonymizedQuery
}
