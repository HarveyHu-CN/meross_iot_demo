package mysql

import (
	"fmt"
	"github.com/albertwidi/sqlt"
	"github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultProtocol = "tcp"
	DefaultPort = 3306
	DefaultTimeout = 3 * time.Second
	DefaultReadTimeout = 3 * time.Second
	DefaultWriteTimeout = 3 * time.Second
	DefaultConnMaxLife = 12 * time.Hour
	DefaultMaxIdleConns = 4
	DefaultMaxOpenConns = 5
	DefaultPacketMaxSize = 4 << 20
	DefaultCollation = "utf8mb4_general_ci"
	DefaultLocale = "UTC"
)

// mysql 总配置
type Config struct {
	// access
	User string
	Password string
	Protocol string
	Address string
	Port int
	DbName string
	// pool
	Timeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	ConnMaxLife time.Duration
	MaxIdleConns int
	MaxOpenConns int
	// other
	PacketMaxSize int
	Collation string
	Locale string
	ParseTime bool
	ColumnWithAlias bool
	RejectReadOnly bool
}

// 获取默认配置
func NewConfig() *Config {
	return &Config{
		Protocol:        DefaultProtocol,
		Port:            DefaultPort,
		Timeout:         DefaultTimeout,
		ReadTimeout:     DefaultReadTimeout,
		WriteTimeout:    DefaultWriteTimeout,
		ConnMaxLife:     DefaultConnMaxLife,
		MaxIdleConns:    DefaultMaxIdleConns,
		MaxOpenConns:    DefaultMaxOpenConns,
		PacketMaxSize:   DefaultPacketMaxSize,
		Collation:       DefaultCollation,
		Locale:          DefaultLocale,
	}
}

func New(c *Config) *sqlt.DB {
	if c == nil {
		panic(fmt.Errorf("mysql config is empty"))
	}
	mc := mysql.NewConfig()
	mc.User = c.User
	mc.Passwd = c.Password
	mc.Net = c.Protocol
	mc.DBName = c.DbName
	mc.Timeout = c.Timeout
	mc.ReadTimeout = c.ReadTimeout
	mc.WriteTimeout = c.WriteTimeout
	mc.MaxAllowedPacket = c.PacketMaxSize
	mc.Collation = c.Collation
	loc, err := time.LoadLocation(c.Locale)
	if err != nil {
		panic(fmt.Errorf("mysql init failed with error: %s\n", err))
	}
	mc.Loc = loc
	mc.ParseTime = c.ParseTime
	mc.ColumnsWithAlias = c.ColumnWithAlias
	mc.RejectReadOnly = c.RejectReadOnly
	addrs := strings.Split(c.Address, ";")
	addrLen := len(addrs)
	DSNs := make([]string, addrLen)
	for i := 0; i < addrLen; i++ {
		mc.Addr = addrs[i] + ":" + strconv.Itoa(c.Port)
		fmt.Printf("%+v\n", mc)
		DSNs[i] = mc.FormatDSN()
	}
	DSN := strings.Join(DSNs, ";")
	fmt.Printf("%s\n", DSN)
	db, err := sqlt.Open("mysql", DSN)
	if err != nil {
		panic(fmt.Errorf("init mysql failed with error: %s\n", err))
	}
	//设置pool option
	db.SetConnMaxLifetime(c.ConnMaxLife)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetMaxOpenConnections(c.MaxOpenConns)
	return db
}


