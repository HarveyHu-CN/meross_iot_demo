package redis

import (
	"fmt"
	"time"
)

const (
	DriverRedigo = "redigo"
	DriverGoRedis = "go-redis"

	DefaultHost = "127.0.0.1"
	DefaultPort = 6379
	DefaultDb = 0
	DefaultTimeout = 3 * time.Second
	DefaultReadTimeout = 3 * time.Second
	DefaultWriteTimeout = 3 * time.Second
	DefaultKeepalive = 2 * time.Minute
	DefaultMaxActiveConns = 5
	DefaultMaxIdleConns = 4
	DefaultIdleTimeout = 3 * time.Hour
	DefaultConnMaxLife = 12 * time.Hour
	DefaultPingOnBorrow = 2 * time.Minute
	DefaultDriver = DriverRedigo
)

type Config struct {
	Driver string
	// net
	Host string
	Port int
	Password string
	Database int
	// conn
	Timeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	Keepalive time.Duration
	// pool
	MaxActiveConns int
	MaxIdleConns int
	IdleTimeout time.Duration
	ConnMaxLife time.Duration
	PingOnBorrow time.Duration
}

type Redis struct {
	conf *Config
	pool Pool
}

func NewConfig() *Config {
	return &Config{
		Driver:			DefaultDriver,
		Host:			DefaultHost,
		Port:			DefaultPort,
		Database:       DefaultDb,
		Timeout:        DefaultTimeout,
		ReadTimeout:    DefaultReadTimeout,
		WriteTimeout:   DefaultWriteTimeout,
		Keepalive:      DefaultKeepalive,
		MaxActiveConns: DefaultMaxActiveConns,
		MaxIdleConns:   DefaultMaxIdleConns,
		IdleTimeout:    DefaultIdleTimeout,
		ConnMaxLife:    DefaultConnMaxLife,
		PingOnBorrow:   DefaultPingOnBorrow,
	}
}

func New(c *Config) *Redis {
	if c == nil {
		panic(fmt.Errorf("redis config is empty"))
	}
	if c.Database < 0 || c.Database > 15 {
		panic(fmt.Errorf("wrong redis config: %+v\n", c))
	}
	if c.Timeout < 0 || c.ReadTimeout < 0 || c.WriteTimeout < 0 || c.Keepalive < 0 {
		panic(fmt.Errorf("wrong redis config: %+v\n", c))
	}
	if c.MaxActiveConns < 0 || c.MaxIdleConns < 0 || c.IdleTimeout < 0 {
		panic(fmt.Errorf("wrong redis config: %+v\n", c))
	}
	if c.ConnMaxLife < 0 || c.PingOnBorrow < 0 {
		panic(fmt.Errorf("wrong redis config: %+v\n", c))
	}
	return &Redis{
		conf: c,
	}
}

func (r *Redis) Pool() Pool {
	if r.pool != nil {
		return r.pool
	}
	driver := r.conf.Driver
	switch driver {
	case DriverRedigo:
		return newRedigoPool(r.conf)
	case DriverGoRedis:
		return nil
	default:
		panic(fmt.Errorf("unsupported redis client driver [%s]\n", driver))
	}
}

func (r *Redis) PubSubConn() (PubSub, error) {
	driver := r.conf.Driver
	switch driver {
	case DriverRedigo:
		return newRedigoPubSubConn(r.conf)
	case DriverGoRedis:
		return nil, nil
	default:
		panic(fmt.Errorf("unsupported redis client driver [%s]\n", driver))
	}
}

func (r *Redis) BlockedConn() (BlockedConn, error) {
	driver := r.conf.Driver
	switch driver {
	case DriverRedigo:
		return newRedigoBlockedConn(r.conf)
	case DriverGoRedis:
		return nil, nil
	default:
		panic(fmt.Errorf("unsupported redis client driver [%s]\n", driver))
	}
}

