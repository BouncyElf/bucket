// Copyright (c) 2025 BouncyElf
// SPDX-License-Identifier: MIT

package bucket

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	BucketEventKey       = "bucket.event.key"
	EventPass      Event = iota
	EventRejected
	EventIPNotFound
)

var (
	ErrInvalidTokenNumber     = wrapErr(errors.New("invalid token number"))
	ErrRefillIntervalTooSmall = wrapErr(errors.New("refill interval too small"))
)

type Event int

type AtomicBucket struct {
	token     int64 // atomic
	updatedAt int64 // UnixNano
}

type Config struct {
	Storage           Storage
	TokenNumber       int64
	RefillMicrosecond int64 // 微秒级精度
	EventHook         gin.HandlerFunc
	WeakRejectionMode bool
}

type Storage interface {
	GetOrCreate(key string, creator func() *AtomicBucket) *AtomicBucket
}

type defaultStorage struct {
	m  cmap.ConcurrentMap
	mu sync.Mutex
}

func (s *defaultStorage) GetOrCreate(key string, creator func() *AtomicBucket) *AtomicBucket {
	if v, ok := s.m.Get(key); ok {
		return v.(*AtomicBucket)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.m.Get(key); ok {
		return v.(*AtomicBucket)
	}

	b := creator()
	s.m.Set(key, b)
	return b
}

func NewDefaultConfig() *Config {
	return &Config{
		Storage: &defaultStorage{
			m: cmap.New(),
		},
		TokenNumber:       10000,
		RefillMicrosecond: 100, // 100μs补充1个token
	}
}

func (conf *Config) Valid() error {
	if conf.TokenNumber < 1 {
		return ErrInvalidTokenNumber
	}
	if conf.RefillMicrosecond < 10 { // 最小10μs间隔
		return ErrRefillIntervalTooSmall
	}
	return nil
}

func BucketHandler(conf *Config) gin.HandlerFunc {
	if err := conf.Valid(); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			handleEvent(conf, c, EventIPNotFound)
			if conf.WeakRejectionMode {
				c.Next()
				return
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		bucket := conf.Storage.GetOrCreate(key, func() *AtomicBucket {
			return newBucket(conf)
		})

		now := time.Now().UnixNano()
		oldUpdated := atomic.LoadInt64(&bucket.updatedAt)
		elapsed := now - oldUpdated

		refillNanos := conf.RefillMicrosecond * 1000
		tokensToAdd := elapsed / refillNanos
		if tokensToAdd > 0 {
			newToken := atomic.AddInt64(&bucket.token, tokensToAdd)
			if newToken > conf.TokenNumber {
				atomic.StoreInt64(&bucket.token, conf.TokenNumber)
			}
			atomic.StoreInt64(&bucket.updatedAt, now)
		}

		if atomic.AddInt64(&bucket.token, -1) < 0 {
			atomic.AddInt64(&bucket.token, 1) // 回滚
			handleEvent(conf, c, EventRejected)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		handleEvent(conf, c, EventPass)
		c.Next()
	}
}

func newBucket(conf *Config) *AtomicBucket {
	if conf == nil {
		return nil
	}
	return &AtomicBucket{
		token:     conf.TokenNumber,
		updatedAt: time.Now().UnixNano(),
	}
}

func handleEvent(conf *Config, c *gin.Context, e Event) {
	c.Set(BucketEventKey, e)
	if conf.EventHook != nil {
		conf.EventHook(c)
	}
}

func wrapErr(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("bucket err: %v", err)
}
