package recentcontact

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xcache/functions"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

//go:embed operation.lua
var operationLua string

func Init(rd *redis.Redis) {
	if err := functions.FunctionLoadReplace(context.Background(), rd, operationLua); err != nil {
		panic(fmt.Errorf("recent contact cache init failed: %w", err))
	}
}

const (
	maxDayMs       = xtime.DaySec * 7 * 1000
	maxCount       = 50
	cleanThreshold = 200
)

// Store stores recent contacts of a user, following the below cache structure:
//
// Max-lengthed sorted set: uidkey -> {{member:targetuid1, score:time}, {member:targetuid2, score:time}, ...},
// with time in millisecond
type Store struct {
	rd *redis.Redis
}

func New(rd *redis.Redis) *Store {
	return &Store{
		rd: rd,
	}
}

const (
	userRecentContactKeyPrefix = "pilot.user.recentcontact."
)

func makeUserRecentContactKey(uid int64) string {
	return userRecentContactKeyPrefix + strconv.FormatInt(uid, 10)
}

type RecentContact struct {
	Uid    int64
	TimeMs int64
}

// 获取最近maxCount个联系人
func (s *Store) GetAll(ctx context.Context, uid int64) ([]RecentContact, error) {
	// 那最近maxDayMs的数据
	key := makeUserRecentContactKey(uid)
	now := time.Now().UnixMicro()

	resp, err := s.rd.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, now-maxDayMs, now, 0, maxCount)
	if err != nil {
		return nil, xerror.Wrapf(err, "recent contact zrange rev failed")
	}

	ret := make([]RecentContact, 0, len(resp))
	for _, r := range resp {
		uid, err := strconv.ParseInt(r.Key, 10, 64)
		if err == nil {
			ret = append(ret, RecentContact{
				Uid:    uid,
				TimeMs: r.Score,
			})
		}
	}

	return ret, nil
}

// 清除maxDayMs前的联系人历史
func (s *Store) CleanExire(ctx context.Context, uid int64) error {
	key := makeUserRecentContactKey(uid)
	cutoff := time.Now().UnixMicro() - maxDayMs

	_, err := functions.FunctionCall(ctx, s.rd, "recent_contact_cleanup",
		[]string{key},
		cleanThreshold, 0, cutoff,
	)
	if err != nil {
		return xerror.Wrapf(err, "recent contact cleanup failed")
	}

	return nil
}

func (s *Store) Append(ctx context.Context, uid int64, targets []int64) error {
	key := makeUserRecentContactKey(uid)
	pipe, err := s.rd.TxPipeline()
	if err != nil {
		return xerror.Wrapf(err, "recent contact tx pipeline failed")
	}

	now := time.Now().UnixMicro()
	slices.Sort(targets)
	for _, target := range targets {
		pipe.ZAdd(ctx, key, redis.Z{
			Member: target,
			Score:  float64(now),
		})
	}

	pipe.Expire(ctx, key, xtime.Week)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return xerror.Wrapf(err, "recent contact tx exec failed")
	}

	return nil
}

func (s *Store) AtomicAppend(ctx context.Context, uid int64, targets []int64) error {
	key := makeUserRecentContactKey(uid)

	score := time.Now().UnixMicro()
	args := make([]any, 0, len(targets)*2+2)
	args = append(args, maxCount, xtime.WeekSec)
	for _, target := range targets {
		args = append(args, score, target)
	}

	_, err := functions.FunctionCall(ctx, s.rd, "recent_contact_append",
		[]string{key},
		args...,
	)
	if err != nil {
		return xerror.Wrapf(err, "recent contact append failed")
	}

	return nil
}
