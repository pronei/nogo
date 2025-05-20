package strategy

import (
	"fmt"
	"time"

	"github.com/pronei/nogo/internal/store"
	structs "github.com/pronei/nogo/shared"
)

type StaticWindow struct {
	unitWrapper func(time.Time) int64
}

func getStaticWindow(unitDispatch func(time.Time) int64) Limiter {
	return &StaticWindow{unitWrapper: unitDispatch}
}

func (l *StaticWindow) Allowed(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) (bool, error) {
	currentTime := l.unitWrapper(time.Now())
	allow, err := evaluate(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) bool {
		// approach -> logs[rule.Duration * (currentTime - rule.Duration) : currentTime].length < rule.Limit
		logCount := len(attrState.Logs)
		for _, subRule := range attrRule.Rates {
			windowSize := subRule.Duration
			windowStart := windowSize * (currentTime / windowSize)
			idx := findWindowStartIndex(attrState.Logs, windowStart)
			if !(idx <= logCount && logCount-idx <= subRule.Limit) {
				return false
			}
		}
		return true
	})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate rules - %w\n", err)
	}
	return allow, nil
}

func (l *StaticWindow) UpdateState(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) error {
	// TODO: find a way to ensure the same timestamp is used for both check & update calls
	currentTime := l.unitWrapper(time.Now())
	return changeState(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) error {
		// purge logs older than maximum of all subRule durations
		var windowSize int64
		for _, rate := range attrRule.Rates {
			if rate.Duration > windowSize {
				windowSize = rate.Duration
			}
		}
		windowStart := windowSize * (currentTime / windowSize)
		idx := findWindowStartIndex(attrState.Logs, windowStart)
		attrState.Logs = append(attrState.Logs[idx:], currentTime)
		attrState.LastUpdated = currentTime
		return nil
	})
}
