package structs

import (
	"github.com/pronei/nogo/internal/enums"
	"github.com/redis/go-redis/v9"
)

// TODO: make it more generic to support different dbs
type RateLimiterConfig struct {
	Namespace           string         `json:"namespace"`
	StrategyConfig      StrategyConfig `json:"strategyConfig"`
	StorageType         enums.Storage  `json:"storageType"`
	RedisConfig         RedisConfig    `json:"redisConfig"`
	InMemoryConfig      InMemoryConfig `json:"inMemoryConfig"`
	ExistingRedisClient *redis.Client
}

type StrategyConfig struct {
	Type     string `json:"type"`
	TimeUnit string `json:"timeUnit"`
}

type RedisConfig struct {
	Host                      string `json:"host"`
	Password                  string `json:"password"`
	ConnectionTimeoutInMillis int    `json:"connTimeoutMs"`
	ReadTimeoutInMillis       int    `json:"readTimeoutMs"`
	WriteTimeoutInMillis      int    `json:"writeTimeoutMs"`
	PoolSize                  int    `json:"poolSize"`
	DB                        int    `json:"dbNo"`
}

type InMemoryConfig struct {
	Expiration      Duration `json:"expiration"`
	CleanupInterval Duration `json:"cleanupInterval"`
}
