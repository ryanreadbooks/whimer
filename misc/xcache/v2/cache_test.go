package v2

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var testRedis *redis.Redis

func TestMain(m *testing.M) {
	testRedis = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	m.Run()
}

type testCacheObj struct {
	Id   int    `redis:"id" mapstructure:"id" json:"id"`
	Name string `redis:"name" mapstructure:"name" json:"name"`
}

func TestJsonMarshal(t *testing.T) {
	var s *testCacheObj
	err := json.Unmarshal([]byte(`{"id":100,"name":"test"}`), &s)
	t.Log(err)
	t.Log(s)
}

func TestHGetAllOrFetch(t *testing.T) {
	Convey("HGetAllOrFetch", t, func() {
		result, err := New[*testCacheObj](testRedis).HGetAllOrFetch(
			context.Background(),
			"__test:hgetall",
			func(ctx context.Context) (*testCacheObj, time.Duration, error) {
				t.Log("fallback")
				return &testCacheObj{Id: 1, Name: "test"}, 0, nil
			},
			WithTTL(time.Minute),
		)
		So(err, ShouldBeNil)
		So(result.Id, ShouldEqual, 1)
		So(result.Name, ShouldEqual, "test")
	})

}

func TestHGetAllOrFetch2(t *testing.T) {
	Convey("HGetAllOrFetch no pointer", t, func() {
		result, err := New[testCacheObj](testRedis).HGetAllOrFetch(
			context.Background(),
			"__test:hgetall_2",
			func(ctx context.Context) (testCacheObj, time.Duration, error) {
				t.Log("fallback")
				return testCacheObj{Id: 2, Name: "test"}, 0, nil
			},
			WithTTL(time.Minute),
		)
		So(err, ShouldBeNil)
		So(result.Id, ShouldEqual, 2)
		So(result.Name, ShouldEqual, "test")
	})
}
