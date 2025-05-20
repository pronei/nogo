package structs

import (
	"github.com/redis/go-redis/v9"
)

// TODO: make it more generic to support different dbs
type RateLimiterConfig struct {
	Namespace           string         `json:"namespace"`
	StrategyConfig      StrategyConfig `json:"strategyConfig"`
	RedisConfig         RedisConfig    `json:"redisConfig"`
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
