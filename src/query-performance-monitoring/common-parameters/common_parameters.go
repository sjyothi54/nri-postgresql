package commonparameters

import (
	"github.com/newrelic/nri-postgresql/src/args"
)

type CommonParameters struct {
	Version   uint64
	Databases string
	Arguments args.ArgumentList
}

func SetCommonParameters(args args.ArgumentList, version uint64, databases string) *CommonParameters {
	return &CommonParameters{
		Version:   version,
		Databases: databases,
		Arguments: args,
	}
}
