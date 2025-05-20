package client

import (
	"context"
	"fmt"

	"github.com/pronei/nogo/internal/cache"
	"github.com/pronei/nogo/internal/enums"
	"github.com/pronei/nogo/internal/helpers"
	"github.com/pronei/nogo/internal/store"
	"github.com/pronei/nogo/internal/strategy"
	structs "github.com/pronei/nogo/shared"
)

type RateLimiter interface {
	Allowed(context.Context, *structs.LimitRequest) (bool, error)
	AllowAndUpdate(context.Context, *structs.LimitRequest) (bool, error)
	UpdateRules(*structs.RuleImport, enums.RuleAction) error
	GetRulesByKeys([]string) map[string]structs.EntityRules
}

type rateLimiter struct {
	ruleCache  *cache.RuleCache
	stateStore store.StateStore
	checker    strategy.Limiter
	logger     helpers.Logger
}

var registry map[string]RateLimiter

func init() {
	registry = make(map[string]RateLimiter)
}

// Create instantiates a rate limiter, ingests rules and adds to the registry
func Create(logger helpers.Logger, config *structs.RateLimiterConfig, importedRules *structs.RuleImport) (RateLimiter, error) {

	checker, err := strategy.FromConfig(&config.StrategyConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create strategizer - %w\n", err)
	}

	var stateStore store.StateStore
	if config.ExistingRedisClient != nil {
		stateStore = store.FromRedisClient(logger, config.ExistingRedisClient, config.Namespace)
	} else {
		stateStore, err = store.NewRedisClient(logger, &config.RedisConfig, config.Namespace)
		if err != nil {
			return nil, fmt.Errorf("Unable to create state store - %w\n", err)
		}
	}

	rl := &rateLimiter{
		ruleCache:  cache.New(),
		stateStore: stateStore,
		checker:    checker,
		logger:     logger,
	}
	if err := rl.ruleCache.SaveRules(importedRules, enums.RuleAdd); err != nil {
		return nil, fmt.Errorf("Unable to ingest rules - %w\n", err)
	}

	registry[config.Namespace] = rl
	return rl, nil
}

// Get a rate limiter from the registry if it has been created before
func Get(namespace string) (RateLimiter, error) {
	rl, exists := registry[namespace]
	if !exists {
		return nil, fmt.Errorf("No ratelimiter client found for namespace - %s\n", namespace)
	}
	return rl, nil
}

// Allowed checks if a request is deemed eligible to process as per the rules defined and the current state
func (rl *rateLimiter) Allowed(ctx context.Context, request *structs.LimitRequest) (bool, error) {

	// NEW - ALL attribute is added for each entity in client request
	AddAllAttributesForAllEntities(request)

	// fetch valid rules for the given request from cache
	rulesInCache := rl.ruleCache.GetValidRules(request)
	if len(rulesInCache) == 0 {
		rl.logger.Info("no rules found in cache for %v\n", request)
		return true, nil
	}

	// create entity state request from cached rules
	stateRequest := store.CreateStateRequest(rulesInCache)

	// fetch the current state from backing store
	stateMap, err := rl.stateStore.GetState(ctx, stateRequest)
	//rl.logger.Info("state map snap - %#v\n", stateMap)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve state - %w\n", err)
	}

	// check if the rules allow the current state to be updated
	pass, err := rl.checker.Allowed(rulesInCache, stateMap)
	if err != nil {
		return false, fmt.Errorf("strategy: pass check failure - %w\n", err)
	}
	if !pass {
		return false, nil
	}

	return true, nil
}

func (rl *rateLimiter) AllowAndUpdate(ctx context.Context, request *structs.LimitRequest) (bool, error) {

	// NEW - ALL attribute is added for each entity in client request
	AddAllAttributesForAllEntities(request)

	// fetch valid rules for the given request from cache
	rulesInCache := rl.ruleCache.GetValidRules(request)
	if len(rulesInCache) == 0 {
		rl.logger.Info("no rules found in cache for %v\n", request)
		return true, nil
	}

	// create entity state request from cached rules
	stateRequest := store.CreateStateRequest(rulesInCache)

	// fetch the current state from backing store
	stateMap, err := rl.stateStore.GetState(ctx, stateRequest)
	//rl.logger.Info("state map snap - %#v\n", stateMap)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve state - %w\n", err)
	}

	// check if the rules allow the current state to be updated
	pass, err := rl.checker.Allowed(rulesInCache, stateMap)
	if err != nil {
		return false, fmt.Errorf("strategy: pass check failure - %w\n", err)
	}
	if !pass {
		return false, nil
	}

	// change the state by incrementing counters/updating log windows depending upon the strategy
	if err := rl.checker.UpdateState(rulesInCache, stateMap); err != nil {
		return false, fmt.Errorf("strategy: update failure - %w\n", err)
	}

	// update the state to the backing store (TODO: can be done is async but what about errors?)
	//rl.logger.Info("state before sending to redis %#v\n", stateMap)
	if err := rl.stateStore.SetState(ctx, stateMap); err != nil {
		return false, fmt.Errorf("failed to store state - %w\n", err)
	}

	return true, nil
}

func (rl *rateLimiter) UpdateRules(update *structs.RuleImport, action enums.RuleAction) error {
	if err := rl.ruleCache.SaveRules(update, action); err != nil {
		return fmt.Errorf("Failed to update rules - %w\n", err)
	}
	return nil
}

func (rl *rateLimiter) GetRulesByKeys(keys []string) map[string]structs.EntityRules {
	return rl.ruleCache.GetRulesForKeys(keys)
}
