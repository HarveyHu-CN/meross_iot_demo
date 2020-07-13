package redis

import (
	"context"
	"time"
)

/* *********************************
 * ********** redis pool ***********
 * *********************************/

type PoolStat struct {
	ActiveCount int
	IdleCount int
}

type Connection interface {
	Close() error
	Error() error
	Do(command string, args ...interface{}) (reply interface{}, err error)
}

type BlockedConn interface {
	DoWithTimeout(timeout time.Duration, command string, args ...interface{}) (reply interface{}, err error)
	Close() error
}

type Pool interface {
	Borrow() (Connection, error)
	BorrowWithContext(ctx context.Context) (Connection, error)
	Close() error
	Stat() *PoolStat
	Pipeline() Pipeline
	Script(int, string) Script
}

/* *********************************
 * ******** redis pipeline *********
 * *********************************/

type Pipeline interface {
	Send(cmd string, args ...interface{})
	Exec(ctx context.Context) (*Replies, error)
}

type cmd struct {
	command string
	args        []interface{}
}

type reply struct {
	reply interface{}
	err   error
}

type Replies struct {
	replies []*reply
}

/* *********************************
 * ********* redis script **********
 * *********************************/

type Script interface {
	Do(ctx context.Context, args ...interface{}) (reply interface{}, err error)
	Hash() string
	Load(ctx context.Context) error
}

/* *********************************
 * ********* redis pubsub **********
 * *********************************/

type SubMsg struct {
	// The originating channel.
	Channel string
	// The message data.
	Data []byte
}

type PsubMsg struct {
	// The matched pattern.
	Pattern string
	// The originating channel.
	Channel string
	// The message data.
	Data []byte
}

type Subscription struct {
	// Kind is "subscribe", "unsubscribe", "psubscribe" or "punsubscribe"
	Kind string
	// The channel that was changed.
	Channel string
	// The current number of subscriptions for connection.
	Count int
}

type Pong struct {
	Data string
}

type PubSub interface {
	Close() error
	Subscribe(channels ...interface{}) error
	Unsubscribe(channels ...interface{}) error
	PSubscribe(channels ...interface{}) error
	PUnsubscribe(channels ...interface{}) error
	Receive() interface{}
	ReceiveWithTimeout(timeout time.Duration) interface{}
	Ping(string) error
}
