package commonutils

import "errors"

const MaxQueryThreshold = 30
const MaxIndividualQueryThreshold = 10
const PublishThreshold = 100
const RandomIntRange = 1000000
const TimeFormat = "20060102150405"
const VersionRegex = "PostgreSQL (\\d+)\\."

var ParseVersionError = errors.New("unable to parse PostgreSQL version from string")
var UnsupportedVersion = errors.New("unsupported PostgreSQL version")
var VersionFetchError = errors.New("no rows returned from version query")

const PostgresVersion12 = 12
const PostgresVersion13 = 13
const PostgresVersion14 = 14
const VersionIndex = 2
