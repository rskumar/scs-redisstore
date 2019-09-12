// Package redisstore is a Redis-based session store for the SCS session package.
//
// Warning: API is not finalized and may change, possibly significantly.
//
// The redisstore package relies on the the popular go-redis/redis Redis client
// (github.com/go-redis/redis).

package redisstore

import (
	"time"
	"github.com/go-redis/redis/v7"
)

// Prefix controls the Redis key prefix. You should only need to change this if there is
// a naming clash.
var Prefix = "scs:session:"

// RedisStore represents the currently configured session session store. It supports
// any client that implements redis.Cmdable interface, ex. redis.Client, redis.ClusterClient etc
type RedisStore struct {
	pool redis.Cmdable
}

// New returns a new RedisStore instance. The pool parameter should be any redis.Cmdable implementation
func New(pool redis.Cmdable) *RedisStore {
	return &RedisStore{pool}
}

// Find returns the data for a given session token from the RedisStore instance. If the session
// token is not found or is expired, the returned exists flag will be set to false.
func (r *RedisStore) Find(token string) (b []byte, exists bool, err error) {
	b, err = r.pool.Get(Prefix + token).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Save adds a session token and data to the RedisStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (r *RedisStore) Save(token string, b []byte, expiry time.Time) error {

	cmdz, err := r.pool.TxPipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.Set(Prefix+token, b, 0)
		pipeliner.ExpireAt(Prefix+token, expiry)
		return nil
	})
	if err != nil {
		return err
	}
	for _, cmd := range cmdz {
		if cmd.Err() != nil {
			return err
		}
	}
	return nil
}

// Delete removes a session token and corresponding data from the ResisStore instance.
func (r *RedisStore) Delete(token string) error {
	err := r.pool.Del(Prefix + token).Err()
	if err != nil {
		return err
	}
	return nil
}

func makeMillisecondTimestamp(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
