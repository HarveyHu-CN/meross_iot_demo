package redis_test

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"meross_iot/library/cache/redis"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testRedisSuite struct {
	suite.Suite
	confDefault *redis.Config
	confPwd *redis.Config
	confIllegal *redis.Config
	confWrongDriver *redis.Config
	mr *miniredis.Miniredis
}

func (s *testRedisSuite) SetupSuite()  {
	s.confDefault = redis.NewConfig()

	err := error(nil)
	s.mr, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
	s.confPwd = redis.NewConfig()
	parts := strings.Split(s.mr.Addr(), ":")
	s.confPwd.Host = parts[0]
	s.confPwd.Port, _ = strconv.Atoi(parts[1])
	s.confPwd.Password = "123456"

	s.confIllegal = redis.NewConfig()
	s.confIllegal.ReadTimeout = -2 * time.Second
	s.confIllegal.MaxIdleConns = -5
	s.confIllegal.ConnMaxLife = -100 * time.Second

	s.confWrongDriver = redis.NewConfig()
	s.confWrongDriver.Driver = "xxxxx"
}

func (s *testRedisSuite) TearDownSuite() {
	s.mr.Close()
}

/*
 * 1. 测试redis包的new方法对config的行为
 */
func (s *testRedisSuite) TestNew() {
	//default config
	assert.NotPanics(s.T(), func() {
		redis.New(s.confDefault)
	})
	//illegal config
	assert.Panics(s.T(), func() {
		redis.New(s.confIllegal)
	})
}

/*
 * 2. 测试Redis实例的Pool方法对config的行为
 */
func (s *testRedisSuite) TestPool() {
	//default config
	r := redis.New(s.confDefault)
	assert.NotPanics(s.T(), func() {
		r.Pool()
	})
	//wrong driver
	r = redis.New(s.confWrongDriver)
	assert.Panics(s.T(), func() {
		r.Pool()
	})
}

/*
 * 3. 测试Redis实例的PubSubConn方法对config的行为
 */
func (s *testRedisSuite) TestPubSubConn() {
	//default config
	r := redis.New(s.confDefault)
	assert.NotPanics(s.T(), func() {
		_, err := r.PubSubConn()
		assert.NoError(s.T(), err)
	})
	//wrong driver
	r = redis.New(s.confWrongDriver)
	assert.Panics(s.T(), func() {
		r.PubSubConn()
	})
}

/*
 * 4. 测试Redis实例的PubSubConn方法对config的行为
 */
func (s *testRedisSuite) TestBlockedConn() {
	//default config
	r := redis.New(s.confDefault)
	assert.NotPanics(s.T(), func() {
		_, err := r.BlockedConn()
		assert.NoError(s.T(), err)
	})
	//wrong driver
	r = redis.New(s.confWrongDriver)
	assert.Panics(s.T(), func() {
		r.BlockedConn()
	})
}

/*
 * 5. 测试redis的auth方法，基于PubSubConn方法
 */
func (s *testRedisSuite) TestRedisAuth() {
	// password正确
	s.mr.RequireAuth("123456")
	r := redis.New(s.confPwd)
	_, err := r.PubSubConn()
	//fmt.Printf("%+v\n", err)
	assert.NoError(s.T(), err)
	//password错误
	s.mr.RequireAuth("xxxxxx")
	_, err = r.PubSubConn()
	//fmt.Printf("%+v\n", err)
	assert.Error(s.T(), err)
}


func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(testRedisSuite))
}


