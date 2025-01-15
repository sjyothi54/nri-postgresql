package global_variables

import "github.com/newrelic/nri-postgresql/src/args"

var Args = args.ArgumentList{}
var SlowQuery string = ""
var BlockingQuery string = ""
var IndividualQuery string = ""
var Version uint64 = 0
var DatabaseString = ""
