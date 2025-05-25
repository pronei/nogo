package cache

import (
	"fmt"

	"github.com/patrickmn/go-cache"
	"github.com/pronei/nogo/internal/enums"
	"github.com/pronei/nogo/internal/helpers"
	structs "github.com/pronei/nogo/shared"
)

type RuleCache struct {
	c *cache.Cache
}

func (rc *RuleCache) SaveRules(imported *structs.RuleImport, action enums.RuleAction) error {
	for _, entity := range imported.EntityRuleMap {
		for _, attribute := range entity.EntityAttributes {
			cacheKey := helpers.FormKey(entity.EntityType, attribute.AttributeType, attribute.AttributeValue)
			switch action {
			case enums.RuleDelete:
				rc.c.Delete(cacheKey)
			case enums.RuleAdd:
				if _, exists := rc.c.Get(cacheKey); exists {
					return fmt.Errorf("duplicate rule exists for %s\n", cacheKey)
				}
				rc.c.Set(cacheKey, attribute, cache.NoExpiration)
			case enums.RuleUpdate:
				if err := rc.c.Replace(cacheKey, attribute, cache.NoExpiration); err != nil {
					return fmt.Errorf("cannot update key - %w", err)
				}
			}
		}
	}
	return nil
}

func (rc *RuleCache) GetValidRules(req *structs.LimitRequest) map[string]structs.EntityRules {
	result := make(map[string]structs.EntityRules)
	for entityName, params := range req.Parameters {
		var attributes []structs.AttributeRule
		for attributeType, attributeValue := range params.AttributesMap {
			ruleKey := helpers.FormKey(params.EntityType, attributeType, attributeValue)
			if val, exists := rc.c.Get(ruleKey); exists {
				attributeRule := val.(structs.AttributeRule)
				attributes = append(attributes, attributeRule)
			}
		}
		if len(attributes) > 0 {
			result[helpers.FormKey(params.EntityType, entityName)] = structs.EntityRules{
				EntityName:       entityName,
				EntityType:       params.EntityType,
				EntityAttributes: attributes,
			}
		}
	}
	return result
}

func (rc *RuleCache) GetRulesForKeys(entityType []string) map[string]structs.EntityRules {
	entityTypes := make(map[string]bool)
	for _, entity := range entityType {
		entityTypes[entity] = true
	}

	result := make(map[string]structs.EntityRules)
	rules := rc.c.Items()
	for k, v := range rules {
		key := helpers.SplitKeys(k)
		if entityTypes[key[0]] {
			attributeRule := v.Object.(structs.AttributeRule)
			if _, exists := result[key[0]]; exists {
				entities := result[key[0]]
				entities.EntityAttributes = append(entities.EntityAttributes, attributeRule)
				result[key[0]] = entities
			} else {
				entities := structs.EntityRules{}
				entities.EntityAttributes = append(entities.EntityAttributes, attributeRule)
				result[key[0]] = entities
			}
		}
	}
	return result
}

func New() *RuleCache {
	return &RuleCache{c: cache.New(cache.NoExpiration, cache.NoExpiration)}
}
