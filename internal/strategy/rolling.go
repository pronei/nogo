package strategy

import (
	"fmt"
	"time"

	"github.com/pronei/nogo/internal/store"
	structs "github.com/pronei/nogo/shared"
)

type SlidingWindow struct {
	unitWrapper func(t time.Time) int64
}

func getSlidingWindow(unitDispatch func(time.Time) int64) Limiter {
	return &SlidingWindow{unitWrapper: unitDispatch}
}

func (l *SlidingWindow) Allowed(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) (bool, error) {
	currentTime := l.unitWrapper(time.Now())
	allow, err := evaluate(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) bool {
		// approach -> logs[currentTime - rule.Duration : currentTime].length < rule.Limit
		logCount := len(attrState.Logs)
		for _, subRule := range attrRule.Rates {
			windowStart := currentTime - subRule.Duration
			idx := findWindowStartIndex(attrState.Logs, windowStart)
			if !(idx <= logCount && logCount-idx < subRule.Limit) {
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

func (l *SlidingWindow) UpdateState(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) error {
	currentTime := l.unitWrapper(time.Now())
	return changeState(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) error {
		// purge logs older than maximum of all subRule durations
		var windowSize int64
		for _, rate := range attrRule.Rates {
			if rate.Duration > windowSize {
				windowSize = rate.Duration
			}
		}
		windowStart := currentTime - windowSize
		idx := findWindowStartIndex(attrState.Logs, windowStart)
		attrState.Logs = append(attrState.Logs[idx:], currentTime)
		attrState.LastUpdated = l.unitWrapper(time.Now())
		return nil
	})
}
