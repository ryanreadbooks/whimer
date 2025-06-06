package dao

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	rd          *redis.Redis
	testSessDao *SessionDao
	ctx         = context.TODO()
)

func TestMain(m *testing.M) {
	rd = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	testSessDao = NewSessionDao(rd)
	m.Run()
}

func TestSessionDao_Create(t *testing.T) {
	err := testSessDao.Create(ctx, &Session{
		Id:             uuid.NewString(),
		Uid:            930495,
		Device:         "test",
		Status:         100,
		Ctime:          time.Now().Unix(),
		LastActiveTime: time.Now().Add(time.Second * 10).Unix(),
		Reside:         gofakeit.IPv4Address(),
		Ip:             gofakeit.IPv4Address(),
	})
	t.Log(err)
}

func TestSessionDao_GetById(t *testing.T) {
	s, err := testSessDao.GetById(ctx, "2b100b68-999b-46e2-a961-7fc6ba304d1d")
	t.Log(err)
	t.Logf("%+v\n", s)

	s, err = testSessDao.GetById(ctx, "non-exists")
	t.Log(err)
	t.Logf("%+v\n", s)
}
