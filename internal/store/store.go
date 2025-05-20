package store

import (
	"context"

	"github.com/pronei/nogo/internal/helpers"
	structs "github.com/pronei/nogo/shared"
)

type StateRequestMap map[string]EntityRequest

type EntityRequest struct {
	Type            string
	Name            string
	AttributeStates []AttributeRequest
}

type AttributeRequest struct {
	Key   string
	Value string
}

type StateMap map[string]EntityState

type EntityState struct {
	EntityType        string                    `json:"entityType"`
	EntityName        string                    `json:"entityName"`
	AttributeStateMap map[string]AttributeState `json:"attributes"`
}

type AttributeState struct {
	Bucket      int64   `json:"bucket"`
	Logs        []int64 `json:"logs"`
	LastUpdated int64   `json:"lastUpdated"`
}

type StateStore interface {
	// NOTE: key -> entity type+name, value -> map of attribute type+value to AttributeState

	GetState(ctx context.Context, req StateRequestMap) (StateMap, error)
	SetState(ctx context.Context, state StateMap) error
}

func CreateStateRequest(rules map[string]structs.EntityRules) StateRequestMap {
	reqMap := make(StateRequestMap)
	for _, entity := range rules {
		entityType := entity.EntityType
		entityName := entity.EntityName

		attrStates := make([]AttributeRequest, len(entity.EntityAttributes))
		for i, rule := range entity.EntityAttributes {
			attrStates[i] = AttributeRequest{
				Key:   rule.AttributeType,
				Value: rule.AttributeValue,
			}
		}
		reqMap[helpers.FormKey(entityType, entityName)] = EntityRequest{
			Type:            entityType,
			Name:            entityName,
			AttributeStates: attrStates,
		}
	}
	return reqMap
}
