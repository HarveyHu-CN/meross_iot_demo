package redis_test

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"meross_iot/library/cache/redis"
	"testing"
	"time"
)

type testAdaptorRedigoSuite struct {
	suite.Suite
	c *redis.Config
	mr *miniredis.Miniredis
}

func (s *testAdaptorRedigoSuite) SetupSuite()  {
	s.c = redis.NewConfig()
	s.c.MaxActiveConns = 3
	s.c.MaxIdleConns = 2
}

func (s *testAdaptorRedigoSuite) TearDownSuite() {

}

/*
 * 1. 测试Pool的borrow方法
 */
func (s *testAdaptorRedigoSuite) TestPoolBorrow()  {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	p := r.Pool()
	stat := p.Stat()
	assrt.Equal(0, stat.ActiveCount)
	assrt.Equal(0, stat.IdleCount)
	conn1, _ := p.Borrow()
	stat = p.Stat()
	assrt.Equal(1, stat.ActiveCount)
	assrt.Equal(0, stat.IdleCount)
	conn2, _ := p.Borrow()
	stat = p.Stat()
	assrt.Equal(2, stat.ActiveCount)
	assrt.Equal(0, stat.IdleCount)
	conn3, _ := p.Borrow()
	stat = p.Stat()
	assrt.Equal(3, stat.ActiveCount)
	assrt.Equal(0, stat.IdleCount)
	conn1.Close()
	stat = p.Stat()
	assrt.Equal(3, stat.ActiveCount)
	assrt.Equal(1, stat.IdleCount)
	conn2.Close()
	stat = p.Stat()
	assrt.Equal(3, stat.ActiveCount)
	assrt.Equal(2, stat.IdleCount)
	conn3.Close()
	stat = p.Stat()
	assrt.Equal(2, stat.ActiveCount)
	assrt.Equal(2, stat.IdleCount)
	// 测试关闭
	p.Close()
	stat = p.Stat()
	assrt.Equal(0, stat.ActiveCount)
	assrt.Equal(0, stat.IdleCount)
}

/*
 * 2. 测试pool用尽
 */
func (s *testAdaptorRedigoSuite) TestPoolExhaustedBorrow()  {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	p := r.Pool()
	ch := make(chan int)
	p.Borrow()
	p.Borrow()
	p.Borrow()
	go func() {
		p.Borrow()
		ch<-1
	}()
	select {
	case <-ch:
		assrt.Fail("Borrowed a connection when pool exhausted")
	case <-time.After(500 * time.Millisecond):
		assrt.Equal(1, 1)
	}
	p.Close()
}

/*
 * 3. 测试BorrowWithContext
 */
func (s *testAdaptorRedigoSuite) TestPoolExhaustedBorrowWithContext()  {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	p := r.Pool()
	ch := make(chan int)
	ctx := context.Background()
	p.BorrowWithContext(ctx)
	p.BorrowWithContext(ctx)
	p.BorrowWithContext(ctx)
	go func() {
		p.BorrowWithContext(ctx)
		ch<-1
	}()
	select {
	case <-ch:
		assrt.Fail("Borrowed a connection with context when pool exhausted")
	case <-time.After(500 * time.Millisecond):
		assrt.Equal(1, 1)
	}
	p.Close()
}

/*
 * 4. 测试pool用尽之后，BorrowWithContext被取消的场景
 */
func (s *testAdaptorRedigoSuite) TestPoolCancelBorrowWithContext() {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	p := r.Pool()
	bctx := context.Background()
	ctx, _ := context.WithTimeout(bctx, 100 * time.Millisecond)
	p.BorrowWithContext(bctx)
	p.BorrowWithContext(bctx)
	p.BorrowWithContext(bctx)
	conn, err := p.BorrowWithContext(ctx)
	assrt.Equal(nil, conn)
	assrt.Error(err)
}

/*
 * 5. 测试conn的Do接口
 */
func (s *testAdaptorRedigoSuite) TestConnectionDo() {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	p := r.Pool()
	bctx := context.Background()
	ctx, _ := context.WithTimeout(bctx, 100 * time.Millisecond)
	conn, _ := p.BorrowWithContext(ctx)
	// SET & GET
	reply, err := conn.Do("SET", "foo", "a")
	assrt.Equal(nil, err)
	assrt.Equal("OK", reply)
	reply, err = conn.Do("GET", "foo")
	assrt.Equal(nil, err)
	assrt.NotEqual("a", reply)
	reply, err = redis.String(reply, err)
	assrt.Equal(nil, err)
	assrt.Equal("a", reply)
	// TTL
	reply, err = conn.Do("Del", "foo")
	assrt.Equal(nil, err)
	assrt.Equal(int64(1), reply)
	reply, err = conn.Do("SET", "foo", "a", "ex", 100)
	assrt.Equal(nil, err)
	assrt.Equal("OK", reply)
	reply, err = conn.Do("TTL", "foo")
	assrt.Equal(nil, err)
	assrt.Equal(int64(100), reply)
	// HASH
	reply, err = conn.Do("Del", "myhash")
	reply, err = conn.Do("HSET", "myhash", "mykey", "myvalue")
	assrt.Equal(nil, err)
	assrt.Equal(int64(1), reply)
	reply, err = conn.Do("HGET", "myhash", "mykey")
	assrt.Equal(nil, err)
	assrt.NotEqual("myvalue", reply)
	reply, err = redis.String(reply, err)
	assrt.Equal(nil, err)
	assrt.Equal("myvalue", reply)

	reply, err = conn.Do("Del", "foo")
	reply, err = conn.Do("Del", "myhash")
	p.Close()
}

/*
 * 6. 测试conn的DoWithTimeout接口
 */
func (s *testAdaptorRedigoSuite) TestConnectionDoWithTimeout() {
	assrt := assert.New(s.T())
	r := redis.New(s.c)
	bc, err := r.BlockedConn()
	assrt.NoError(err)
	assrt.NotNil(bc)
	//block timeout
	reply, err := bc.DoWithTimeout(100 * time.Millisecond, "BLpop", "test:list", 100)
	assrt.Equal(nil, reply)
	assrt.Error(err)
	bc = nil
	//block and get data
	go func() {
		select {
		case <- time.After(100 * time.Millisecond):
			p := r.Pool()
			c, _ := p.Borrow()
			c.Do("Lpush", "test:list", "aaa")
			p.Close()
		}
	}()
	bc, err = r.BlockedConn()
	assrt.NoError(err)
	assrt.NotNil(bc)
	reply, err = bc.DoWithTimeout(200 * time.Millisecond, "BLpop", "test:list", 200)
	assrt.NoError(err)
	replies, err := redis.Strings(reply, err)
	assrt.NoError(err)
	assrt.Equal("test:list", replies[0])
	assrt.Equal("aaa", replies[1])
	// 测试blockConn的close方法
	err = bc.Close()
	assrt.NoError(err)
}

func TestAdaptorRedigoSuite(t *testing.T) {
	suite.Run(t, new(testAdaptorRedigoSuite))
}


