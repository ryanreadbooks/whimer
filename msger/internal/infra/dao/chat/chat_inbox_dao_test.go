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
