package p2p

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInboxDao_BatchCreate(t *testing.T) {
	Convey("TestInboxDao_BatchCreate", t, func() {
		data1 := &InboxMsg{
			UserId: 100,
			ChatId: 19,
			MsgId:  190,
			Status: 1,
		}
		data2 := &InboxMsg{
			UserId: 200,
			ChatId: 19,
			MsgId:  190,
			Status: 0,
		}
		err := inboxDao.BatchCreate(ctx, []*InboxMsg{data1, data2})
		So(err, ShouldBeNil)
	})
}
