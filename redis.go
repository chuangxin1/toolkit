package toolkit

import (
	"time"

	"github.com/go-redis/redis"
)

// RedisConfig config
type RedisConfig struct {
	Addr         string
	Password     string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	PoolSize     int
}

// RedisClusterConfig redis cluster configure
type RedisClusterConfig struct {
	// A seed list of host:port addresses of cluster nodes.
	Addrs []string

	// The maximum number of retries before giving up. Command is retried
	// on network errors and MOVED/ASK redirects.
	// Default is 16.
	MaxRedirects int

	// Enables read-only commands on slave nodes.
	ReadOnly bool
	// Allows routing read-only commands to the closest master or slave node.
	RouteByLatency bool

	//OnConnect func(*Conn) error

	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	Password        string

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// PoolSize applies per cluster node and not for the whole cluster.
	PoolSize           int
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
}

var (
	redisConfig        RedisConfig
	redisClusterConfig RedisClusterConfig
)

// RedisCache define
type RedisCache struct {
	c  *redis.Client
	cc *redis.ClusterClient
}

// SetRedisConfig set
func SetRedisConfig(cfg RedisConfig) {
	redisConfig = cfg
}

// SetRedisClusterConfig set
func SetRedisClusterConfig(cfg RedisClusterConfig) {
	redisClusterConfig = cfg
}

// NewRedisCache new RedisCache object
func NewRedisCache() *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:         redisConfig.Addr,
		Password:     redisConfig.Password,
		DialTimeout:  redisConfig.DialTimeout,
		ReadTimeout:  redisConfig.ReadTimeout,
		WriteTimeout: redisConfig.WriteTimeout,
		PoolSize:     redisConfig.PoolSize,
		PoolTimeout:  redisConfig.PoolTimeout,
	})

	client.Ping()
	return &RedisCache{c: client}
}

// NewRedisClusterCache new RedisCluster object
func NewRedisClusterCache() *RedisCache {
	var config redis.ClusterOptions

	config.Addrs = redisClusterConfig.Addrs
	config.MaxRedirects = redisClusterConfig.MaxRedirects
	config.ReadOnly = redisClusterConfig.ReadOnly
	config.RouteByLatency = redisClusterConfig.RouteByLatency

	config.MaxRetries = redisClusterConfig.MaxRetries
	config.MinRetryBackoff = redisClusterConfig.MinRetryBackoff
	config.MaxRetryBackoff = redisClusterConfig.MaxRetryBackoff
	config.Password = redisClusterConfig.Password

	config.DialTimeout = redisClusterConfig.DialTimeout
	config.ReadTimeout = redisClusterConfig.ReadTimeout
	config.WriteTimeout = redisClusterConfig.WriteTimeout

	config.PoolSize = redisClusterConfig.PoolSize
	config.PoolTimeout = redisClusterConfig.PoolTimeout
	config.IdleTimeout = redisClusterConfig.IdleTimeout
	config.IdleCheckFrequency = redisClusterConfig.IdleCheckFrequency

	client := redis.NewClusterClient(&config)

	client.Ping()
	return &RedisCache{cc: client}
}

// Get get value from cache
func (c RedisCache) Get(key string) (string, error) {
	return c.c.Get(key).Result()
}

// GetCluster get value from cluster cache
func (c RedisCache) GetCluster(key string) (string, error) {
	return c.cc.Get(key).Result()
}

// Set set key-value to cache
func (c RedisCache) Set(key, value string, expiration time.Duration) error {
	return c.c.Set(key, value, expiration).Err()
}

// SetCluster set key-value to cache
func (c RedisCache) SetCluster(
	key, value string,
	expiration time.Duration) error {
	return c.cc.Set(key, value, expiration).Err()
}

// Subscribe subscribe message
func (c RedisCache) Subscribe(
	channels string,
	cb func(channel string, message string, err error)) {
	pubsub := c.c.Subscribe(channels)
	defer pubsub.Close()

	var (
		msg *redis.Message
		err error
	)
	msg, err = pubsub.ReceiveMessage()
	for err == nil {
		cb(msg.Channel, msg.Payload, nil)
		msg, err = pubsub.ReceiveMessage()
	}

	cb("", "", err)

	return
}

// SubscribeCluster subscribe cluster message
func (c RedisCache) SubscribeCluster(
	channels string,
	cb func(channel string, message string, err error)) {
	pubsub := c.cc.Subscribe(channels)
	defer pubsub.Close()

	var (
		msg *redis.Message
		err error
	)
	msg, err = pubsub.ReceiveMessage()
	for err == nil {
		cb(msg.Channel, msg.Payload, nil)
		msg, err = pubsub.ReceiveMessage()
	}

	cb("", "", err)

	return
}

// Publish publish message
func (c RedisCache) Publish(channel, message string) error {
	return c.c.Publish(channel, message).Err()
}

// PublishCluster publish message
func (c RedisCache) PublishCluster(channel, message string) error {
	return c.cc.Publish(channel, message).Err()
}
