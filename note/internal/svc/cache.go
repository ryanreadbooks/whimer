package svc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"
)

const (
	noteCacheKey = "note:note:%d"
)

func getNoteCacheKey(nid uint64) string {
	return fmt.Sprintf(noteCacheKey, nid)
}

func CacheGetNote(ctx context.Context, nid uint64) (*noterepo.Model, error) {
	res, err := infra.Cache().GetCtx(ctx, getNoteCacheKey(nid))
	if err != nil {
		return nil, err
	}

	var ret noterepo.Model
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func CacheDelNote(ctx context.Context, nid uint64) error {
	_, err := infra.Cache().HdelCtx(ctx, getNoteCacheKey(nid))
	return err
}

func CacheSetNote(ctx context.Context, model *noterepo.Model) error {
	content, err := json.Marshal(model)
	if err != nil {
		return err
	}

	ttl := xtime.Week + xtime.JitterDuration(2*xtime.Hour)
	err = infra.Cache().SetexCtx(ctx, getNoteCacheKey(model.Id), utils.Bytes2String(content), int(ttl))
	return err
}
