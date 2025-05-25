package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pronei/nogo/internal/constants"
)

type Logger interface {
	Info(args ...any)
	Debug(args ...any)
	Error(args ...any)
	Warn(args ...any)
	Fatal(args ...any)
	Panic(args ...any)
}

func GetTimeInDurationWithError(timeVal int, timeUnit string) (time.Duration, error) {
	if !Contains(constants.ValidTimeUnits, timeUnit) {
		return 0, fmt.Errorf("invalid timeunit %s passed\n", timeUnit)
	}
	timeoutStr := strconv.Itoa(timeVal) + timeUnit
	return time.ParseDuration(timeoutStr)
}

func Contains(arr []string, key string) bool {
	for _, s := range arr {
		if s == key {
			return true
		}
	}
	return false
}

func FormKey(s ...string) string {
	return strings.Join(s, constants.KeyDelimiter)
}

func SplitKeys(s string) []string {
	return strings.Split(s, constants.KeyDelimiter)
}

// ParseKey creates a new key from the indices specified
// An empty string is returned if there are no/invalid indices
func ParseKey(s string, idx ...int) string {
	if len(idx) == 0 {
		return ""
	}
	var newKey string
	tokens := strings.Split(s, constants.KeyDelimiter)
	for _, i := range idx {
		if i < 0 || i >= len(tokens) {
			return ""
		}
		newKey += tokens[i] + constants.KeyDelimiter
	}
	return newKey[:len(newKey)-1]
}
