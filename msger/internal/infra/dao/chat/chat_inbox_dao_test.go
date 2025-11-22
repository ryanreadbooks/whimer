package chat

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChatInboxDao_Create(t *testing.T) {
	Convey("TestChatInboxDao_Create", t, func() {
		uid := rand.Int63n(100)
		chatId := uuid.MaxUUID()
		err := testChatInboxDao.Create(t.Context(), &ChatInboxPO{
			Uid:    uid,
			ChatId: chatId,
			Mtime:  time.Now().Unix(),
			Ctime:  time.Now().Unix(),
		})
		So(err, ShouldBeNil)
	})
}

func TestChatInboxDao_BatchUpdateLastMsgId(t *testing.T) {
	Convey("TestChatInboxDao_BatchUpdateLastMsgId", t, func() {
		err := testChatInboxDao.BatchUpdateLastMsgId(t.Context(),
			uuid.NewUUID(),
			[]int64{100, 200, 300},
			uuid.NewUUID(), time.Now().Unix())
		So(err, ShouldBeNil)
	})
}

func TestChatInboxDao_SetLastReadMsgId(t *testing.T) {
	Convey("TestChatInboxDao_SetLastReadMsgId", t, func() {
		uid := rand.Int63n(100000)
		chatId := uuid.NewUUID()
		err := testChatInboxDao.Create(t.Context(), &ChatInboxPO{
			Uid:           uid,
			ChatId:        chatId,
			Mtime:         time.Now().Unix(),
			Ctime:         time.Now().Unix(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.MaxUUID(),
			UnreadCount:   10,
		})
		So(err, ShouldBeNil)

		mtime := time.Now().UnixMicro()
		err = testChatInboxDao.SetLastReadMsgId(t.Context(), uid, chatId, mtime)
		So(err, ShouldBeNil)

		// check
		got, err := testChatInboxDao.GetByUidChatId(t.Context(), uid, chatId)
		So(err, ShouldBeNil)
		So(got.LastReadMsgId, ShouldEqual, got.LastMsgId)
		So(got.LastReadTime, ShouldEqual, mtime)
		So(got.UnreadCount, ShouldEqual, 0)
	})
}

func TestChatInboxDao_DecrUnreadCount(t *testing.T) {
	Convey("TestChatInboxDao_SetLastReadMsgId", t, func() {
		uid := rand.Int63n(100000)
		chatId := uuid.NewUUID()
		err := testChatInboxDao.Create(t.Context(), &ChatInboxPO{
			Uid:           uid,
			ChatId:        chatId,
			Mtime:         time.Now().Unix(),
			Ctime:         time.Now().Unix(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.MaxUUID(),
			UnreadCount:   10,
		})
		So(err, ShouldBeNil)

		err = testChatInboxDao.DecrUnreadCount(t.Context(), uid, chatId, time.Now().Unix())
		So(err, ShouldBeNil)
	})
}

func TestChatInboxDao_PageList(t *testing.T) {
	Convey("TestChatInboxDao_PageList", t, func() {
		uid := rand.Int31n(10)
		for i := range 8 {
			_ = i
			testChatInboxDao.Create(t.Context(), &ChatInboxPO{
				Uid:           int64(uid),
				ChatId:        uuid.NewUUID(),
				Mtime:         time.Now().Unix(),
				Ctime:         time.Now().UnixMicro(),
				LastMsgId:     uuid.NewUUID(),
				LastReadMsgId: uuid.NewUUID(),
				UnreadCount:   rand.Int63n(10),
				Status:        model.ChatInboxStatusNormal,
				IsPinned:      model.ChatInboxPinState(rand.Int31n(2)),
			})

			time.Sleep(time.Millisecond * 50)
		}

		gots, err := testChatInboxDao.PageList(t.Context(), int64(uid), math.MaxInt64, 5)
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 5)
	})
}

func TestChatInboxDao_PageListWithPinned(t *testing.T) {
	Convey("TestChatInboxDao_PageListWithPinned", t, func() {
		gots, err := testChatInboxDao.PageListWithPinned(t.Context(), 8, 10, 3)
		So(err, ShouldBeNil)
		for _, g := range gots {
			t.Log(g)
		}
	})
}

func TestChatInboxDao_PageListWithUnPinned(t *testing.T) {
	Convey("TestChatInboxDao_PageListWithUnPinned", t, func() {
		gots, err := testChatInboxDao.PageListWithUnPinned(t.Context(), 8, 10, 3)
		So(err, ShouldBeNil)
		for _, g := range gots {
			t.Log(g)
		}
	})
}
