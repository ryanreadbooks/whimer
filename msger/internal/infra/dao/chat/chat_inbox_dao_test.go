package chat

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
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
