package commonutils

import (
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func FetchVersionSpecificSlowQueries(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12:
		return queries.SlowQueriesForV12, nil
	case version >= PostgresVersion13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchSlowAndIndividualQueriesPgStat(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12:
		return queries.SlowQueryPgStatV12, nil
	case version >= PostgresVersion13:
		return queries.SlowQueryPgStatV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificBlockingQueries(version uint64, isRds bool) (string, error) {
	switch {
	case version == PostgresVersion12, version == PostgresVersion13:
		return queries.BlockingQueriesForV12AndV13, nil
	case version >= PostgresVersion14 && isRds:
		return queries.BlockingQueriesForV14AndAboveQueryMatch, nil
	case version >= PostgresVersion14 && !isRds:
		return queries.BlockingQueriesForV14AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificIndividualQueries(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12:
		return queries.IndividualQuerySearchV12, nil
	case version > PostgresVersion12:
		return queries.IndividualQuerySearchV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchSupportedWaitEvents(cp *commonparameters.CommonParameters) (string, error) {
	if cp.IsRds {
		return queries.WaitEventsFromPgStatActivity, nil
	} else {
		return queries.WaitEvents, nil
	}
}
