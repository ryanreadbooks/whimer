package note

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xstring"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const defaultNumNoteInteractStatList uint32 = 6 // do not modify this value

const (
	// number of key
	noteLikeCountStatSetKey     = "pilot.note.stats.like_count.set"
	noteCommentCountStatSetKey  = "pilot.note.stats.comment_count.set"
	noteLikeCountStatKeyTmpl    = "pilot.note.stats.like_count.list.%d"
	noteCommentCountStatKeyTmpl = "pilot.note.stats.comment_count.list.%d"
)

// Note cache stat representation states:
//
// sorted set: {{list.0, score0}, {list.1, score1}, {list.2, score2}}
//
// list:
//
//	list.0: [{"nid": "xxx", "inc": 1}, {"nid": "xxx", "inc": 1}]
//	list.1: [{"nid": "xxx", "inc": -1}, {"nid": "xxx", "inc": -1}]
//	list.2: [{"nid": "xxx", "inc": 1}, {"nid": "xxx", "inc": -1}]
type StatStore struct {
	rd              *redis.Redis
	likeStatKeys    []string
	commentStatKeys []string
}

func NewStatStore(rd *redis.Redis, numOfList uint32) *StatStore {
	if numOfList == 0 {
		numOfList = defaultNumNoteInteractStatList
	}
	s := &StatStore{
		rd: rd,
	}
	for idx := range numOfList {
		s.likeStatKeys = append(s.likeStatKeys, fmt.Sprintf(noteLikeCountStatKeyTmpl, idx))
		s.commentStatKeys = append(s.commentStatKeys, fmt.Sprintf(noteCommentCountStatKeyTmpl, idx))
	}

	return s
}

type NoteInteractStatType string

const (
	NoteLikeCountStat    NoteInteractStatType = "note_like"
	NoteCommentCountStat NoteInteractStatType = "note_comment"
)

type NoteStatRepr struct {
	Type   NoteInteractStatType `json:"-"`
	NoteId string               `json:"nid"` // note_id
	Inc    int64                `json:"inc"` // increment
}

// 数据先行写入redis
func (b *StatStore) Add(ctx context.Context,
	statType NoteInteractStatType, stat NoteStatRepr,
) error {
	hasher := fnv.New32a()
	hasher.Write(xstring.AsBytes(stat.NoteId))
	slotIdx := int(hasher.Sum32()) % int(defaultNumNoteInteractStatList)
	reprByte, err := json.Marshal(&stat)
	if err != nil {
		return xerror.Wrapf(err, "marshal note stat repr failed").WithCtx(ctx)
	}

	var (
		listKey   string
		listValue = xstring.FromBytes(reprByte)
		setKey    string
	)

	switch statType {
	case NoteLikeCountStat:
		listKey = b.likeStatKeys[slotIdx]
		setKey = noteLikeCountStatSetKey
	case NoteCommentCountStat:
		listKey = b.commentStatKeys[slotIdx]
		setKey = noteCommentCountStatSetKey
	default:
		return xerror.ErrArgs.Msgf("unsupported note stat type: %s", statType)
	}

	err = b.rd.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		now := time.Now().UnixMicro() // 使用时间作为score模拟队列 实现FIFO
		p.ZAdd(ctx, setKey, redis.Z{Score: float64(now), Member: listKey})
		p.LPush(ctx, listKey, listValue)
		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "lpush to %s failed, body: %s", listKey, listValue)
	}

	return nil
}

func (b *StatStore) ConsumeLikeCount(ctx context.Context, want int) ([]NoteStatRepr, error) {
	return b.consumeByType(ctx, NoteLikeCountStat, int64(want))
}

func (b *StatStore) ConsumeCommentCount(ctx context.Context, want int) ([]NoteStatRepr, error) {
	return b.consumeByType(ctx, NoteCommentCountStat, int64(want))
}

// want：每次获取sorted set中的元素的个数
func (b *StatStore) consumeByType(ctx context.Context, statType NoteInteractStatType, want int64) ([]NoteStatRepr, error) {
	pipe, err := b.rd.TxPipeline()
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	var setKey string
	switch statType {
	case NoteLikeCountStat:
		setKey = noteLikeCountStatSetKey
	case NoteCommentCountStat:
		setKey = noteCommentCountStatSetKey
	}

	// 这里多个命令不是原子的理论不影响
	// 如果zpopmin后又有内容插入了list 要么这次处理 要么留到下次处理

	var shouldCompensate bool
	pipe.ZPopMin(ctx, setKey, want)
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	zCmd, ok := cmds[0].(*goredis.ZSliceCmd)
	if !ok {
		return nil, xerror.Wrap(xerror.ErrInternal)
	}

	zRes, err := zCmd.Result()
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	defer func() {
		if shouldCompensate {
			concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
				Name: "note_cache.consume.compensate",
				Job: func(ctx context.Context) error {
					if err := b.rd.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
						for idx := range len(zRes) {
							zRes[idx].Score = float64(time.Now().UnixMicro()) // 重置时间
						}
						p.ZAdd(ctx, setKey, zRes...)
						return nil
					}); err != nil {
						xlog.Msg("note cache compensate err occurred").Err(err).Errorx(ctx)
					}

					return nil
				},
			})
		}
	}()

	listKeys := []string{}
	for _, zcmd := range zRes {
		if listKey, ok := zcmd.Member.(string); ok {
			listKeys = append(listKeys, listKey)
		}
	}

	totalItems := make([]string, 0, 16)
	// rpop all listKeys
	for _, listKey := range listKeys {
		listLen, err := b.rd.LlenCtx(ctx, listKey)
		if err != nil {
			xlog.Msg("consume by type llen failed").
				Extras("list_key", listKey).Err(err).Errorx(ctx)
			shouldCompensate = true
			continue
		}

		// rpop listLen elements from listKey list
		listItems, err := b.rd.RpopCountCtx(ctx, listKey, listLen)
		if err != nil {
			xlog.Msg("consume by type rpop failed").
				Extras("list_key", listKey, "len", listLen).
				Err(err).Errorx(ctx)
			shouldCompensate = true
			continue
		}

		totalItems = append(totalItems, listItems...)
	}

	ret := make([]NoteStatRepr, 0, len(totalItems))
	for _, item := range totalItems {
		var itemStat NoteStatRepr
		err = json.Unmarshal(xstring.AsBytes(item), &itemStat)
		if err == nil {
			itemStat.Type = statType
			ret = append(ret, itemStat)
		} else {
			xlog.Msg("consume by type unmarshal data failed").Err(err).Errorx(ctx)
		}
	}

	return ret, nil
}
