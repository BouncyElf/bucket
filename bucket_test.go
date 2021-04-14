package bucket

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_defaultStorage(t *testing.T) {
	tt := struct {
		name    string
		setData map[string]string
		getData map[string]string
	}{
		name: "happy path",
		setData: map[string]string{
			"k1": "v1",
		},
		getData: map[string]string{
			"k1": "v1",
			"k":  "",
			"":   "",
		},
	}

	s := new(defaultStorage)
	assert.NotPanics(t, func() {
		for k, v := range tt.setData {
			s.Set(k, v)
		}
	}, tt.name)
	assert.NotPanics(t, func() {
		for k, v := range tt.getData {
			assert.Equal(t, v, s.Get(k), tt.name)
		}
	}, tt.name)
}

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() {
		New()
	})
}

func Test_eventHappen(t *testing.T) {
	tests := []struct {
		name  string
		conf  *Config
		c     *gin.Context
		event string
		err   error
	}{
		{
			name: "happy path(empty)",
			conf: DefaultConfig,
		},
	}
	for _, tt := range tests {
		assert.NotPanics(t, func() {
			eventHappen(tt.conf, tt.c, tt.event, tt.err)
		}, tt.name)
	}
}
