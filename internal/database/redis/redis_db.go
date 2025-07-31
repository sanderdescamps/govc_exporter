package redis_db

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type DB struct {
	client   *redis.Client
	password string
	addr     string
	dbIndex  int
}

type RedisDbOption func(db *DB)

func WithPassword(password string) RedisDbOption {
	return func(db *DB) {
		db.password = password
	}
}

func WithDBIndex(i int) RedisDbOption {
	return func(db *DB) {
		db.dbIndex = i
	}
}

func NewDB(addr string, options ...RedisDbOption) *DB {
	db := &DB{
		addr:     addr,
		password: "",
		dbIndex:  0,
	}
	for _, o := range options {
		o(db)
	}

	return db
}

func (db *DB) Connect(ctx context.Context) error {
	if db.client != nil {
		return nil
	}

	db.client = redis.NewClient(&redis.Options{
		Addr:     db.addr,
		Password: db.password, // no password set
		DB:       db.dbIndex,  // use default DB
	})

	return nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	db.client = nil
	return nil
}

func (db *DB) SetObj(ctx context.Context, key string, path string, obj interface{}, ttl time.Duration) error {
	db.Connect(ctx)

	status := db.client.JSONSet(ctx, key, path, obj)
	if status.Err() != nil {
		return status.Err()
	}

	if ttl != 0 {
		_, err := db.client.Expire(ctx, key, ttl).Result()
		if err != nil {
			return fmt.Errorf("failed to set expiration of object: %v", err)
		}
	}
	return nil
}

func (db *DB) SetHost(ctx context.Context, host *mo.HostSystem, ttl time.Duration) error {
	db.Connect(ctx)

	err := db.SetObj(ctx, host.Self.Value, "HostSystem", host, ttl)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetAllHostIter(ctx context.Context) iter.Seq2[*mo.HostSystem, error] {
	db.Connect(ctx)
	redisIter := db.client.Scan(ctx, 0, "HostSystem:*", 0).Iterator()
	return func(yield func(*mo.HostSystem, error) bool) {
		for redisIter.Next(ctx) {
			var host *mo.HostSystem
			err := json.Unmarshal([]byte(redisIter.Val()), host)
			if !yield(host, err) {
				return
			}
		}
	}
}

func (db *DB) GetAllHostRefs(ctx context.Context) []types.ManagedObjectReference {
	result := []types.ManagedObjectReference{}
	for host := range db.GetAllHostIter(ctx) {
		result = append(result, host.Self)
	}
	return result
}
