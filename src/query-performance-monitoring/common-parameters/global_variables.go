package commonparameters

import (
	"github.com/newrelic/nri-postgresql/src/args"
)

type CommonParameters struct {
	Version        uint64
	DatabaseString string
	Arguments      args.ArgumentList
}

func SetCommonParameters(args args.ArgumentList, version uint64, databaseString string) *CommonParameters {
	return &CommonParameters{
		Version:        version,
		DatabaseString: databaseString,
		Arguments:      args,
	}
}
