package client

import (
	"fmt"

	"github.com/pronei/nogo/internal/constants"
	structs "github.com/pronei/nogo/shared"
)

func AddAllAttributesForEntity(req *structs.LimitRequest, entityType, entityName string) error {
	if reqEntity, exists := req.Parameters[entityName]; exists &&
		reqEntity.EntityType == entityType {
		reqEntity.AttributesMap[constants.AllAttribute] = constants.AllAttribute
		return nil
	} else {
		return fmt.Errorf("could not find entity name with corresponding type in request params - %#v\n", req.Parameters)
	}
}

func AddAllAttributesForAllEntities(req *structs.LimitRequest) {
	for _, params := range req.Parameters {
		params.AttributesMap[constants.AllAttribute] = constants.AllAttribute
	}
}

func AddAllEntityWithAttributes(req *structs.LimitRequest, attrMap map[string]string) error {
	if len(attrMap) == 0 {
		return fmt.Errorf("cannot attach ALL entity with empty attributes")
	}
	req.Parameters[constants.AllEntity] = structs.EntityParameters{
		EntityType:    constants.AllEntity,
		AttributesMap: attrMap,
	}
	return nil
}
