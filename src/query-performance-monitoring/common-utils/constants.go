package commonutils

import "errors"

const MAX_QUERY_THRESHOLD = 30
const MAX_INDIVIDUAL_QUERY_THRESHOLD = 10
const PUBLISH_THRESHOLD = 100
const RANDOM_INT_RANGE = 1000000
const TIME_FORMAT = "20060102150405"
const VERSION_REGEX = "PostgreSQL (\\d+)\\."

var PARSE_VERSION_ERROR = errors.New("unable to parse PostgreSQL version from string")
var UNSUPPORTED_VERSION = errors.New("unsupported PostgreSQL version")

const POSTGRES_VERSION_12 = 12
const POSTGRES_VERSION_13 = 13
const POSTGRES_VERSION_14 = 14
const VERSION_INDEX = 2
