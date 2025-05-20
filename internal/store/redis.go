package store

import (
	"context"
	"fmt"

	"github.com/pronei/nogo/internal/constants"
	"github.com/pronei/nogo/internal/helpers"
	protobuf "github.com/pronei/nogo/internal/proto"
	structs "github.com/pronei/nogo/shared"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type redisClient struct {
	client    *redis.Client
	keyPrefix string
	logger    helpers.Logger
}

func NewRedisClient(logger helpers.Logger, opts *structs.RedisConfig, namespace string) (StateStore, error) {
	db := opts.DB
	readTimeout, _ := helpers.GetTimeInDurationWithError(opts.ReadTimeoutInMillis, constants.MilliSecond)
	writeTimeout, _ := helpers.GetTimeInDurationWithError(opts.WriteTimeoutInMillis, constants.MilliSecond)
	serverOpts := &redis.Options{
		Addr:         opts.Host,
		Password:     opts.Password,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     opts.PoolSize,
		DB:           db,
	}
	client := redis.NewClient(serverOpts)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("Could not connect to Redis @ %s, DB - %d, error - %v\n", "", db, err.Error())
	}
	logger.Info("Connected to redis: %s\n", pong)
	return &redisClient{client: client, keyPrefix: namespace, logger: logger}, nil
}

func FromRedisClient(logger helpers.Logger, client *redis.Client, namespace string) StateStore {
	return &redisClient{client: client, keyPrefix: namespace, logger: logger}
}

func (r *redisClient) GetState(ctx context.Context, req StateRequestMap) (StateMap, error) {
	stateMap := make(StateMap)

	// index based lookup - invariant -> command results in the pipeline are in order
	var hashKeys []string

	pipe := r.client.Pipeline()
	for _, entityReq := range req {
		key := helpers.FormKey(entityReq.Type, entityReq.Name)
		attrKeys := getAttributeKeys(&entityReq)
		if err := pipe.HMGet(ctx, r.keyPrefix+key, attrKeys...).Err(); err != nil {
			return nil, fmt.Errorf("failed to build HMGet pipeline - %w\n", err)
		}
		hashKeys = append(hashKeys, key)
	}

	commands, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HMGet pipeline - %w\n", err)
	}

	// assert len(hashKeys) == len(commands)
	for hashIdx, cmd := range commands {
		cmdResult, err := cmd.(*redis.SliceCmd).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to type assert result from pipeline - %w\n", err)
		}

		attributeStateMap := make(map[string]AttributeState)
		hashKey := hashKeys[hashIdx]

		for attrIdx, result := range cmdResult {
			// no state yet
			if result == nil {
				continue
			}

			attributeReq := req[hashKey].AttributeStates[attrIdx]
			attributeKey := helpers.FormKey(attributeReq.Key, attributeReq.Value)
			attributeState, err := getAttributeFromProtoBytes([]byte(result.(string)))
			if err != nil {
				return nil, fmt.Errorf("failed to parse state for attribute %s - %w\n", attributeKey, err)
			}
			if _, exists := attributeStateMap[attributeKey]; !exists {
				attributeStateMap[attributeKey] = *attributeState
			}
		}

		if _, exists := stateMap[hashKey]; !exists && len(attributeStateMap) > 0 {
			stateMap[hashKey] = EntityState{
				EntityType:        helpers.ParseKey(hashKey, 0),
				EntityName:        helpers.ParseKey(hashKey, 1),
				AttributeStateMap: attributeStateMap,
			}
		}
	}

	return stateMap, nil
}

func (r *redisClient) SetState(ctx context.Context, state StateMap) error {
	pipe := r.client.Pipeline()
	for _, entity := range state {
		key := helpers.FormKey(entity.EntityType, entity.EntityName)
		protoAttrMap := make(map[string]interface{})
		for attrKey, attrVal := range entity.AttributeStateMap {
			if _, exists := protoAttrMap[attrKey]; !exists {
				bytes, err := getProtoBytesForAttribute(&attrVal, attrKey)
				if err != nil {
					// TODO: break on single marshalling failure
					r.logger.Warn("skipping key %s in HSet for %s - %s\n", attrKey, key, err.Error())
					continue
				}
				protoAttrMap[attrKey] = bytes
			}
		}
		if err := pipe.HSet(ctx, r.keyPrefix+key, protoAttrMap).Err(); err != nil {
			return fmt.Errorf("pipeline error in HSet for key %s - %w\n", key, err)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute HSet pipeline - %w\n", err)
	}
	return nil
}

func getProtoBytesForAttribute(attribute *AttributeState, key string) ([]byte, error) {
	bytes, err := proto.Marshal(&protobuf.AttributeState{
		Bucket:      attribute.Bucket,
		Logs:        attribute.Logs,
		LastUpdated: attribute.LastUpdated,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to marshal key %s into proto - %w\n", key, err)
	}
	return bytes, nil
}

func getAttributeFromProtoBytes(b []byte) (*AttributeState, error) {
	attributeProto := &protobuf.AttributeState{}
	if err := proto.Unmarshal(b, attributeProto); err != nil {
		return nil, err
	}
	return &AttributeState{
		Bucket:      attributeProto.Bucket,
		Logs:        attributeProto.Logs,
		LastUpdated: attributeProto.LastUpdated,
	}, nil
}

func getAttributeKeys(entityReq *EntityRequest) []string {
	var keys []string
	for _, attribute := range entityReq.AttributeStates {
		keys = append(keys, helpers.FormKey(attribute.Key, attribute.Value))
	}
	return keys
}
