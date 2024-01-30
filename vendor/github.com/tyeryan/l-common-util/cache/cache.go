package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/wire"
	"github.com/tyeryan/l-common-util/config"
	logutil "github.com/tyeryan/l-protocol/log"
	"github.com/vmihailenco/msgpack"
)

var (
	WireSet = wire.NewSet(
		ProvideRedisConfig,
		ProvideRedisClient,
		ProvideCacheClient,
	)

	log = logutil.GetLogger("cache-client")

	clientMutex    = &sync.Mutex{}
	clientInstance *redis.ClusterClient

	mutex    = &sync.Mutex{}
	instance *CacheClient
)

// RedisConfig stores Redis connection details
type RedisConfig struct {
	Host         string `configstruct:"RedisConfig_Host" configdefault:"you-redis-cluster.youapp.svc"`
	Port         int    `configstruct:"RedisConfig_Port" configdefault:"6379"`
	Password     string `configstruct:"RedisConfig_Password" configdefault:""`
	MaxRedirects int    `configstruct:"RedisConfig_MaxRedirects" configdefault:"0"`
	// FIXME: DB index support
}

type DistributedCache interface {
	Get(key string, v interface{}) error
	Set(key string, expiration time.Duration, v interface{}) error
	Del(key string) error
	GetClient() *redis.ClusterClient
	UpdateWithNewRedisClusterClient() error
}

// CacheClient cache client
type CacheClient struct {
	client *redis.ClusterClient
	cnf    *RedisConfig
}

// ProvideRedisConfig redis config provider
func ProvideRedisConfig(ctx context.Context, configStore config.ConfigStore) (*RedisConfig, error) {
	cnf := &RedisConfig{}
	if err := configStore.GetConfig(cnf); err != nil {
		return nil, err
	}
	return cnf, nil
}

func ProvideRedisClient(ctx context.Context, cnf *RedisConfig) (*redis.ClusterClient, error) {
	log.Debugw(ctx, "ProvideRedisClient begin")
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if clientInstance != nil {
		return clientInstance, nil
	}

	address := fmt.Sprintf("%s:%d", cnf.Host, cnf.Port)

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{address},
		Password:     cnf.Password,
		MaxRedirects: cnf.MaxRedirects,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	clientInstance = client
	log.Debugw(ctx, "ProvideRedisClient end")
	return client, nil
}

// ProvideCacheClient cache client provider
func ProvideCacheClient(ctx context.Context,
	redisClient *redis.ClusterClient,
	cnf *RedisConfig) (DistributedCache, error) {
	log.Debugw(ctx, "ProvideCacheClient begin")
	mutex.Lock()
	defer mutex.Unlock()
	if instance != nil {
		return instance, nil
	}

	instance = &CacheClient{
		client: redisClient,
		cnf:    cnf,
	}
	log.Debugw(ctx, "ProvideCacheClient end")
	return instance, nil
}

// GetClient get redis client
func (c *CacheClient) GetClient() *redis.ClusterClient {
	return c.client
}

// Get get object from cache server
func (c *CacheClient) Get(key string, v interface{}) error {
	ctx := context.Background()
	err := c.get(key, v)
	if err != nil && strings.HasSuffix(err.Error(), "i/o timeout") {
		log.Errore(ctx, "timeout error with get", err)
		err = c.UpdateWithNewRedisClusterClient()
		if err == nil {
			err = c.get(key, v)
		}
	}
	return err
}

func (c *CacheClient) UpdateWithNewRedisClusterClient() error {
	address := fmt.Sprintf("%s:%d", c.cnf.Host, c.cnf.Port)
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{address},
		Password:     c.cnf.Password,
		MaxRedirects: c.cnf.MaxRedirects,
	})

	_, err := client.Ping().Result()

	if err != nil {
		return err
	}

	c.client = client
	return nil
}

// Set save object in cache server
func (c *CacheClient) Set(key string, expiration time.Duration, v interface{}) error {
	ctx := context.Background()
	err := c.set(key, expiration, v)
	if err != nil && strings.HasSuffix(err.Error(), "i/o timeout") {
		log.Errore(ctx, "timeout error with set", err)
		err = c.UpdateWithNewRedisClusterClient()
		if err == nil {
			err = c.set(key, expiration, v)
		}
	}
	return err
}

// Set save object in cache server
func (c *CacheClient) Del(key string) error {
	ctx := context.Background()
	err := c.del(key)
	if err != nil && strings.HasSuffix(err.Error(), "i/o timeout") {
		log.Errore(ctx, "timeout error with del", err)
		err = c.UpdateWithNewRedisClusterClient()
		if err == nil {
			err = c.del(key)
		}
	}
	return err
}

// Get get object from cache server
func (c *CacheClient) get(key string, v interface{}) error {
	cacheBytes, err := c.GetClient().Get(key).Bytes()
	if err != nil {
		return err
	}

	if err := msgpack.Unmarshal(cacheBytes, v); err != nil {
		return err
	}

	return nil
}

// Set save object in cache server
func (c *CacheClient) set(key string, expiration time.Duration, v interface{}) error {
	cacheBytes, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return c.GetClient().Set(key, cacheBytes, expiration).Err()
}

// Set save object in cache server
func (c *CacheClient) del(key string) error {
	return c.GetClient().Del(key).Err()
}
