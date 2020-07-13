package redis

import (
	"context"
	redigo "github.com/gomodule/redigo/redis"
	"strconv"
	"time"
)

var ErrPoolExhausted = redigo.ErrPoolExhausted

type dialFunc func() (redigo.Conn, error)
type borrowFunc func(redigo.Conn, time.Time) error

// 生成TestOnBorrow函数
func genTestOnBorrowFunc(span time.Duration) borrowFunc {
	return func(conn redigo.Conn, t time.Time) error {
		if time.Since(t) < span {
			return nil
		}
		_, err := conn.Do("PING")
		return err
	}
}

// 生成Dial函数
func genDialFunc(c *Config) dialFunc {
	return func() (conn redigo.Conn, err error) {
		address := c.Host + ":" + strconv.Itoa(c.Port)
		dialOptions := ([]redigo.DialOption)(nil)
		dialOptions = append(dialOptions, redigo.DialConnectTimeout(c.Timeout))
		dialOptions = append(dialOptions, redigo.DialReadTimeout(c.ReadTimeout))
		dialOptions = append(dialOptions, redigo.DialWriteTimeout(c.WriteTimeout))
		dialOptions = append(dialOptions, redigo.DialKeepAlive(c.Keepalive))
		dialOptions = append(dialOptions, redigo.DialPassword(c.Password))
		dialOptions = append(dialOptions, redigo.DialDatabase(c.Database))
		conn, err = redigo.Dial("tcp", address, dialOptions...)
		return
	}
}

func newRedigoPool(c *Config) *redigoPool {
	// idle不能超过active
	if c.MaxIdleConns > c.MaxActiveConns {
		c.MaxIdleConns = c.MaxActiveConns
	}
	pool := &redigo.Pool{}
	pool.IdleTimeout = c.IdleTimeout
	pool.MaxConnLifetime = c.ConnMaxLife
	pool.MaxActive = c.MaxActiveConns
	pool.MaxIdle = c.MaxIdleConns
	pool.Wait = true
	if c.PingOnBorrow != 0 {
		pool.TestOnBorrow = genTestOnBorrowFunc(c.PingOnBorrow)
	}
	pool.Dial = genDialFunc(c)
	return &redigoPool{rp:pool}
}

func newRedigoPubSubConn(c *Config) (*redigo.PubSubConn, error) {
	dial := genDialFunc(c)
	conn, err := dial()
	if err != nil {
		return nil, err
	}
	return &redigo.PubSubConn{Conn:conn}, nil
}

func newRedigoBlockedConn(c *Config) (*redigoConn, error) {
	dial := genDialFunc(c)
	conn, err := dial()
	if err != nil {
		return nil, err
	}
	return &redigoConn{rc:conn}, nil
}



/* *********************************
 * ******** interface Pool *********
 * *********************************/

type redigoPool struct {
	rp *redigo.Pool
}

/*
 * 如果pool用尽，该函数会直接返回一个错误
 */
func (p *redigoPool) Borrow() (Connection, error) {
	rc := p.rp.Get()
	if rc.Err() != nil {
		return nil, rc.Err()
	}
	return &redigoConn{rc}, nil
}

func (p *redigoPool) BorrowWithContext(ctx context.Context) (Connection, error) {
	rc, err := p.rp.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	conn := &redigoConn{
		rc,
	}
	return conn, nil
}

func (p *redigoPool) Close() error {
	return p.rp.Close()
}

func (p *redigoPool) Stat() *PoolStat {
	ps := p.rp.Stats()
	return &PoolStat{
		ActiveCount: ps.ActiveCount,
		IdleCount:   ps.IdleCount,
	}
}

func (p *redigoPool) Pipeline() Pipeline {
	return &redigoPipeline{pool:p.rp}
}

func (p *redigoPool) Script(keyCnt int, src string) Script {
	rs := redigo.NewScript(keyCnt, src)
	return &redigoScript{
		p.rp,
		rs,
	}
}

/* *********************************
 * ****** interface Connection *****
 * *********************************/

type redigoConn struct {
	rc redigo.Conn
}

func (c *redigoConn) Close() error {
	return c.rc.Close()
}

func (c *redigoConn) Error() error {
	return c.rc.Err()
}

func (c *redigoConn) Do(command string, args ...interface{}) (interface{}, error) {
	return c.rc.Do(command, args...)
}

/*
 * redigo的conn当timeout之后，会将底层连接关闭
 */
func (c *redigoConn) DoWithTimeout(t time.Duration, cmd string, args ...interface{}) (interface{}, error) {
	return redigo.DoWithTimeout(c.rc, t, cmd, args...)
}

/* *********************************
 * ******* interface pipeline ******
 * *********************************/

type redigoPipeline struct {
	pool *redigo.Pool
	cmds []*cmd
}

func (p *redigoPipeline) Send(command string, keyAndArg ...interface{})  {
	p.cmds = append(p.cmds, &cmd{command:command, args:keyAndArg})
}

func (p *redigoPipeline) Exec(ctx context.Context) (*Replies, error) {
	n := len(p.cmds)
	if n == 0 {
		return &Replies{}, nil
	}
	c, err := p.pool.GetContext(ctx)
	defer c.Close()
	if err != nil {
		p.cmds = p.cmds[:0]
		return nil, err
	}
	for len(p.cmds) > 0 {
		cmd := p.cmds[0]
		p.cmds = p.cmds[1:]
		if err := c.Send(cmd.command, cmd.args...); err != nil {
			p.cmds = p.cmds[:0]
			return nil, err
		}
	}
	if err = c.Flush(); err != nil {
		p.cmds = p.cmds[:0]
		return nil, err
	}
	rps := make([]*reply, 0, n)
	for i := 0; i < n; i++ {
		rp, err := c.Receive()
		rps = append(rps, &reply{reply: rp, err: err})
	}
	rs := &Replies{
		replies: rps,
	}
	return rs, nil
}

/* *********************************
 * ******* interface script ********
 * *********************************/

type redigoScript struct {
	pool *redigo.Pool
	script *redigo.Script
}

func (s *redigoScript) Do(ctx context.Context, keysAndArgs ...interface{}) (interface{}, error) {
	c, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	v, err := s.script.Do(c, keysAndArgs...)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *redigoScript) Hash() string {
	return s.script.Hash()
}

func (s *redigoScript) Load(ctx context.Context) error {
	c, err := s.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	err = s.script.Load(c)
	if err != nil {
		return err
	}
	return nil
}
/* *********************************
 * ******* interface pubsub ********
 * *********************************/

// 直接使用redigo.PubSubConn满足接口


