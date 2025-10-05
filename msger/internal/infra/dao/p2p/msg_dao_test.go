package p2p

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMsgDao_Create(t *testing.T) {
	Convey("TestMsgDao_Create", t, func() {
		err := messageDao.Create(ctx, &MsgPO{
			MsgId:    1,
			SenderId: 100,
			ChatId:   10,
			MsgType:  1,
			Content:  "data",
			Status:   1,
			Seq:      199,
			Utime:    time.Now().Unix(),
		})
		So(err, ShouldBeNil)
	})
}
