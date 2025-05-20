package strategy

import (
	"fmt"

	"github.com/pronei/nogo/internal/enums"
	"github.com/pronei/nogo/internal/store"
	structs "github.com/pronei/nogo/shared"
)

type Limiter interface {
	Allowed(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) (bool, error)
	UpdateState(ruleMap map[string]structs.EntityRules, stateMap store.StateMap) error
}

func FromConfig(config *structs.StrategyConfig) (Limiter, error) {
	timeDispatch := timeDispatcherMap[config.TimeUnit]
	switch enums.GetStrategy(config.Type) {
	case enums.StrategyRolling:
		return getSlidingWindow(timeDispatch), nil
	case enums.StrategyStatic:
		return getStaticWindow(timeDispatch), nil
	case enums.StrategyFixedBucket:
		return getFixedBucket(timeDispatch), nil
	default:
		return nil, fmt.Errorf("no strategy found for type %s", config.Type)
	}
}
