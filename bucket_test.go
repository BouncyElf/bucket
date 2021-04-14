package bucket

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
)

func Test_defaultStorage_Set(t *testing.T) {
	type fields struct {
		m cmap.ConcurrentMap
	}
	type args struct {
		key string
		val string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		s := &defaultStorage{
			m: tt.fields.m,
		}
		s.Set(tt.args.key, tt.args.val)
	}
}

func Test_defaultStorage_Get(t *testing.T) {
	type fields struct {
		m cmap.ConcurrentMap
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		s := &defaultStorage{
			m: tt.fields.m,
		}
		if got := s.Get(tt.args.key); got != tt.want {
			t.Errorf("%q. defaultStorage.Get() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_defaultSerializer_Marshal(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		d       defaultSerializer
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		d := defaultSerializer{}
		got, err := d.Marshal(tt.args.data)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. defaultSerializer.Marshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. defaultSerializer.Marshal() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_defaultSerializer_Unmarshal(t *testing.T) {
	type args struct {
		bytes    []byte
		receiver interface{}
	}
	tests := []struct {
		name    string
		d       defaultSerializer
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		d := defaultSerializer{}
		if err := d.Unmarshal(tt.args.bytes, tt.args.receiver); (err != nil) != tt.wantErr {
			t.Errorf("%q. defaultSerializer.Unmarshal() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := New(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. New() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBucket(t *testing.T) {
	type args struct {
		conf *Config
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := Bucket(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Bucket() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_newBucket(t *testing.T) {
	type args struct {
		conf *Config
	}
	tests := []struct {
		name string
		args args
		want *bucket
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := newBucket(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. newBucket() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
