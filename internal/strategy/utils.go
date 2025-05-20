package strategy

import (
	"fmt"
	"sort"
	"time"

	"github.com/pronei/nogo/internal/constants"
	"github.com/pronei/nogo/internal/helpers"
	"github.com/pronei/nogo/internal/store"
	structs "github.com/pronei/nogo/shared"
)

// Used to find the appropriate dispatch function for the specified unit
var timeDispatcherMap = make(map[string]func(t time.Time) int64)

func init() {
	timeDispatcherMap[constants.NanoSecond] = time.Time.UnixNano
	timeDispatcherMap[constants.MicroSecond] = time.Time.UnixMicro
	timeDispatcherMap[constants.MilliSecond] = time.Time.UnixMilli
	timeDispatcherMap[constants.Second] = time.Time.Unix
}

// Used to find the first log index of the currently valid window as per windowStart
func findWindowStartIndex(logs []int64, windowStart int64) int {
	return sort.Search(len(logs), func(i int) bool {
		return logs[i] >= windowStart
	})
}

// evaluate iterates over the rule map (entity->attributes) and calls stateChecker
// on states for which ruleKey=entityType:entityName:attrType:attrName match
func evaluate(ruleMap map[string]structs.EntityRules, stateMap store.StateMap,
	stateChecker func(*structs.AttributeRule, *store.AttributeState) bool) (bool, error) {

	for entityKey, rule := range ruleMap {
		if state, exists := stateMap[entityKey]; exists {

			if state.EntityType != rule.EntityType || state.EntityName != rule.EntityName {
				return false, fmt.Errorf("incorrect entity comparison E1 (rule) - %s:%s, E2 (state) - %s:%s\n",
					rule.EntityType, rule.EntityName, state.EntityType, state.EntityName)
			}

			for _, attrRule := range rule.EntityAttributes {
				key := helpers.FormKey(attrRule.AttributeType, attrRule.AttributeValue)
				if attrState, exists := state.AttributeStateMap[key]; exists {
					if !stateChecker(&attrRule, &attrState) {
						return false, nil
					}
				}
			}
		}
	}
	return true, nil
}

// changeState iterates over the rule map (entity->attributes) and calls stateUpdater
// on both existing and new states, modifying them in place. New states have the default struct value.
func changeState(ruleMap map[string]structs.EntityRules, stateMap store.StateMap,
	stateUpdater func(*structs.AttributeRule, *store.AttributeState) error) error {

	for entityKey, entityRule := range ruleMap {

		entityState, exists := stateMap[entityKey]
		if !exists {
			entityState = store.EntityState{
				EntityType:        entityRule.EntityType,
				EntityName:        entityRule.EntityName,
				AttributeStateMap: make(map[string]store.AttributeState),
			}
		}

		for _, attrRule := range entityRule.EntityAttributes {
			attrKey := helpers.FormKey(attrRule.AttributeType, attrRule.AttributeValue)
			attrState := entityState.AttributeStateMap[attrKey]
			if err := stateUpdater(&attrRule, &attrState); err != nil {
				return err
			}
			entityState.AttributeStateMap[attrKey] = attrState
		}

		stateMap[entityKey] = entityState
	}

	return nil
}
