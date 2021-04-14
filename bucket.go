package bucket

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
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
	Storage Storage
	// TODO:

	// serialization, default use json
	Serializer Serializer

	// TokenNumber token number per bucket
	TokenNumber int

	// BucketFillDuration bucket fill duration
	BucketFillDuration time.Duration

	// EventHook is the hook after error or rejected
	EventHook gin.HandlerFunc
}

type bucket struct {
	token     int
	updatedAt time.Time
}

type defaultSerializer struct{}

func (defaultSerializer) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (defaultSerializer) Unmarshal(bytes []byte, receiver interface{}) error {
	return json.Unmarshal(bytes, receiver)
}

// TODO: add default config, use concurrent-map or sync.Map
func Bucket(conf *Config) gin.HandlerFunc {
	if conf == nil {
		panic("Bucket: Missing Config")
	}
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			c.Set(ErrKey, ErrIpNotFound)
			c.Set(EventKey, EventError)
			conf.EventHook(c)
			return
		}
		v := conf.Storage.Get(key)
		b := new(bucket)
		if v == "" {
			b = newBucket(conf)
		}
		err := conf.Serializer.Unmarshal([]byte(v), b)
		if err != nil {
			c.Set(ErrKey, ErrUnmarshalError)
			c.Set(EventKey, EventError)
			conf.EventHook(c)
			return
		}
		if time.Now().After(b.updatedAt.Add(conf.BucketFillDuration)) {
			b.token = conf.TokenNumber
			b.updatedAt = time.Now()
		}
		if b.token <= 0 {
			c.Set(EventKey, EventRejected)
			conf.EventHook(c)
			return
		}
		b.token--
		bs, err := conf.Serializer.Marshal(b)
		if err != nil {
			c.Set(ErrKey, ErrMarshalError)
			c.Set(EventKey, EventError)
			conf.EventHook(c)
			return
		}
		conf.Storage.Set(key, string(bs))
		conf.Set(EventKey, EventPass)
		conf.EventHook(c)
		c.Next()
	}
}

func newBucket(conf *Config) *bucket {
	return &bucket{
		token:     conf.TokenNumber,
		updatedAt: time.Now(),
	}
}
