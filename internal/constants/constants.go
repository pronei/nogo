package constants

const AllEntity = "ALL"
const AllAttribute = "ALL"

const NanoSecond = "ns"
const MicroSecond = "us"
const MilliSecond = "ms"
const Second = "s"
const Minute = "m"
const Hour = "h"

var ValidTimeUnits = []string{NanoSecond, MicroSecond, MilliSecond, Second, Minute, Hour}

// WARN - change with caution, ensure namespace separation does not use this
const KeyDelimiter = ":"
