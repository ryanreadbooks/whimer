package recentcontact

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	testRedis *redis.Redis
	testStore *Store
)

func TestMain(m *testing.M) {
	testRedis = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	Init(testRedis)
	testStore = New(testRedis)

	m.Run()
}

func TestBasic(t *testing.T) {
	uid := int64(1)
	targets := []int64{}
	for range 30 {
		targets = append(targets, rand.Int63n(100))
	}

	ctx := context.Background()
	err := testStore.Append(ctx, uid, targets)
	if err != nil {
		t.Fatal(err)
	}

	gots, err := testStore.GetAll(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(gots)
}

func TestLuaFn(t *testing.T) {
	uid := int64(1)
	targets := []int64{}
	for range 30 {
		targets = append(targets, rand.Int63n(10000))
	}

	ctx := context.Background()
	err := testStore.AtomicAppend(ctx, uid, targets)
	if err != nil {
		t.Fatal(err)
	}

	gots, err := testStore.GetAll(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(gots)
}

func TestNow(t *testing.T) {
	n := time.Now().UnixMicro()
	t.Log(n)
	t.Log(n - maxDayMs)

	gots, err := testStore.GetAll(t.Context(), 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(gots)
}
