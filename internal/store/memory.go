package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/patrickmn/go-cache"
	"github.com/pronei/nogo/internal/helpers"
	structs "github.com/pronei/nogo/shared"
)

type memoryClient struct {
	lock   sync.Mutex
	c      *cache.Cache
	logger helpers.Logger
}

func NewMemoryClient(logger helpers.Logger, opts *structs.InMemoryConfig) StateStore {
	internal := cache.New(opts.Expiration.ToStd(), opts.CleanupInterval.ToStd())
	// this is helpful to offload less frequently accessed keys to a slower store
	internal.OnEvicted(func(k string, v interface{}) {
		eType, eName := helpers.ParseKey(k, 0), helpers.ParseKey(k, 1)
		logger.Info("dropping state for entity %v with type %v", eName, eType)
	})
	return &memoryClient{
		c:      internal,
		logger: logger,
	}
}

func (mc *memoryClient) GetState(_ context.Context, req StateRequestMap) (StateMap, error) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	stateMap := make(StateMap)

	for entityKey, entityReq := range req {
		attrStateMap := make(map[string]AttributeState)

		for _, attrReq := range entityReq.AttributeStates {
			attrKey := helpers.FormKey(attrReq.Key, attrReq.Value)
			stateKey := helpers.FormKey(entityKey, attrKey)
			if val, exists := mc.c.Get(stateKey); exists {
				attrStateMap[attrKey] = val.(AttributeState)
			} else {
				return stateMap, fmt.Errorf("no state found for key - %v", stateKey)
			}
		}

		if _, exists := stateMap[entityKey]; !exists && len(attrStateMap) > 0 {
			stateMap[entityKey] = EntityState{
				EntityType:        entityReq.Type,
				EntityName:        entityReq.Name,
				AttributeStateMap: attrStateMap,
			}
		}
	}

	return stateMap, nil
}

func (mc *memoryClient) SetState(_ context.Context, state StateMap) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	for entityKey, entityState := range state {
		for attrKey, attrState := range entityState.AttributeStateMap {
			key := helpers.FormKey(entityKey, attrKey)
			mc.c.Set(key, attrState, cache.NoExpiration)
		}
	}

	return nil
}
