package dao

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
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
		Id:             "8b4372e0-6786-4fd6-9ba7-81351c72b765",
		Uid:            930495,
		Device:         "test",
		Status:         "active",
		Ctime:          time.Now().Unix(),
		LastActiveTime: time.Now().Add(time.Second * 10).Unix(),
		Reside:         gofakeit.IPv4Address(),
		Ip:             gofakeit.IPv4Address(),
	})
	t.Log(err)

	err = testSessDao.Create(ctx, &Session{
		Id:             "abc927ad-6786-4fd6-9ba7-81351c72b765",
		Uid:            930495,
		Device:         "web",
		Status:         "active",
		Ctime:          time.Now().Unix(),
		LastActiveTime: time.Now().Add(time.Second * 10).Unix(),
		Reside:         gofakeit.IPv4Address(),
		Ip:             gofakeit.IPv4Address(),
	})
	t.Log(err)
}

func TestSessionDao_GetById(t *testing.T) {
	s, err := testSessDao.GetById(ctx, "8b4372e0-6786-4fd6-9ba7-81351c72b765")
	t.Log(err)
	t.Logf("%+v\n", s)

	s, err = testSessDao.GetById(ctx, "non-exists")
	t.Log(err)
	t.Logf("%+v\n", s)
}

func TestSessionDao_DeleteById(t *testing.T) {
	err := testSessDao.DeleteById(ctx, "8b4372e0-6786-4fd6-9ba7-81351c72b765")
	t.Log(err)

	err = testSessDao.DeleteById(ctx, "non-exists")
	t.Log(err)
}

func TestSessionDao_GetByUid(t *testing.T) {
	sesses, err := testSessDao.GetByUid(ctx, 930495)
	t.Log(err)
	for _, s := range sesses {
		t.Log(s)
	}
}

func TestSessionDao_DeleteByUid(t *testing.T) {
	err := testSessDao.DeleteByUid(ctx, 930495)
	t.Log(err)
}

func TestSessionDao_UpdateStatus(t *testing.T) {
	err := testSessDao.UpdateStatus(ctx, "8b4372e0-6786-4fd6-9ba7-81351c72b765", "noactive")
	t.Log(err)
}

func TestSessionDao_UpdateLastActiveTime(t *testing.T) {
	err := testSessDao.UpdateLastActiveTime(ctx, "8b4372e0-6786-4fd6-9ba7-81351c72b765", time.Now().Unix())
	t.Log(err)
}
