package common_package

import "github.com/newrelic/go-agent/v3/newrelic"

var ArgsGlobal = ""
var ArgsApplication = ""
var NewrelicApp = newrelic.Application{}

var Txn *newrelic.Transaction = nil
