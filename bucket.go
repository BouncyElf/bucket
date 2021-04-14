package bucket

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	ErrKey        = "bucket.error.key"
	EventKey      = "bucket.event.key"
	EventError    = "bucket.event.error"
	EventRejected = "bucket.event.reject"
	EventPass     = "bucket.event.pass"
)

var (
	ErrIpNotFound     = errors.New("ip not found")
	ErrUnmarshalError = errors.New("unmarshal error")
	ErrMarshalError   = errors.New("marshal error")
	ErrLimited        = errors.New("rate limited")

	DefaultConfig = &Config{
		Storage:            new(defaultStorage),
		Serializer:         new(defaultSerializer),
		TokenNumber:        10,
		BucketFillDuration: 500 * time.Millisecond,
	}

	once = new(sync.Once)
)

type Storage interface {
	Set(key, val string)
	Get(key string) (val string)
}

type Serializer interface {
	Marshal(data interface{}) ([]byte, error)
	Unmarshal(data []byte, receiver interface{}) error
}

type Config struct {
	// Storage, default use concurrent map
	Storage Storage

	// serialization, default use json
	Serializer Serializer

	// TokenNumber token number per bucket
	TokenNumber int

	// BucketFillDuration bucket fill duration
	BucketFillDuration time.Duration

	// EventHook is the hook after error or rejected
	EventHook gin.HandlerFunc
}

type BucketData struct {
	Token     int
	UpdatedAt time.Time
}

type defaultStorage struct {
	m cmap.ConcurrentMap
}

func (s *defaultStorage) Set(key, val string) {
	once.Do(func() {
		if s.m == nil {
			s.m = cmap.New()
		}
	})
	s.m.Set(key, val)
}

func (s *defaultStorage) Get(key string) string {
	once.Do(func() {
		if s.m == nil {
			s.m = cmap.New()
		}
	})
	if v, ok := s.m.Get(key); ok {
		res, _ := v.(string)
		return res
	}
	return ""
}

type defaultSerializer struct{}

func (defaultSerializer) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (defaultSerializer) Unmarshal(bytes []byte, receiver interface{}) error {
	return json.Unmarshal(bytes, receiver)
}

func New() gin.HandlerFunc {
	return Bucket(DefaultConfig)
}

func Bucket(conf *Config) gin.HandlerFunc {
	if conf == nil {
		panic("Bucket: Missing Config")
	}
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			eventHappen(conf, c, EventError, ErrIpNotFound)
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		v := conf.Storage.Get(key)
		b := newBucket(conf)
		if v != "" {
			err := conf.Serializer.Unmarshal([]byte(v), b)
			if err != nil {
				eventHappen(conf, c, EventError, ErrUnmarshalError)
				c.AbortWithError(http.StatusInternalServerError, ErrUnmarshalError)
				return
			}
		}
		if time.Now().After(b.UpdatedAt.Add(conf.BucketFillDuration)) {
			b.Token = conf.TokenNumber
			b.UpdatedAt = time.Now()
		}
		if b.Token <= 0 {
			eventHappen(conf, c, EventRejected, nil)
			c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
			return
		}
		b.Token--
		bs, err := conf.Serializer.Marshal(b)
		if err != nil {
			eventHappen(conf, c, EventError, ErrMarshalError)
			c.AbortWithError(http.StatusInternalServerError, ErrMarshalError)
			return
		}
		conf.Storage.Set(key, string(bs))
		eventHappen(conf, c, EventPass, nil)
		c.Next()
	}
}

func eventHappen(conf *Config, c *gin.Context, event string, err error) {
	if conf == nil {
		panic("Bucket: Missing Config")
	}
	if event != "" {
		c.Set(EventKey, event)
	}
	if err != nil {
		c.Set(ErrKey, err)
	}
	if h := conf.EventHook; h != nil {
		h(c)
	}
}

func newBucket(conf *Config) *BucketData {
	return &BucketData{
		Token:     conf.TokenNumber,
		UpdatedAt: time.Now(),
	}
}
