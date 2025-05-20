package strategy

import (
	"fmt"
	"math"
	"time"

	"github.com/pronei/nogo/internal/store"
	structs "github.com/pronei/nogo/shared"
)

type FixedBucket struct {
	unitWrapper func(time time.Time) int64
}

func getFixedBucket(unitDispatch func(time.Time) int64) Limiter {
	return &FixedBucket{unitWrapper: unitDispatch}
}

func (l *FixedBucket) Allowed(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) (bool, error) {
	currentTime := l.unitWrapper(time.Now())
	allow, err := evaluate(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) bool {
		tokens := attrState.Bucket
		bucketRule := attrRule.Bucket
		tokensToAdd := int64(math.Round(float64(bucketRule.Refill) * (float64(currentTime-attrState.LastUpdated) / float64(bucketRule.Duration))))
		tokens = min(bucketRule.Maximum, tokens+tokensToAdd)
		return tokens-bucketRule.Cost >= 0
	})
	if err != nil {
		return false, fmt.Errorf("could not evaluate - %w\n", err)
	}
	return allow, nil
}

func (l *FixedBucket) UpdateState(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) error {
	currentTime := l.unitWrapper(time.Now())
	return changeState(ruleMap, stateMap, func(attrRule *structs.AttributeRule, attrState *store.AttributeState) error {
		attrState.Bucket -= attrRule.Bucket.Cost
		attrState.LastUpdated = currentTime
		return nil
	})
}
