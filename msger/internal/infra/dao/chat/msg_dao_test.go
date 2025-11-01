package chat

import (
	"testing"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMsgDao_BatchGetByIds(t *testing.T) {
	Convey("BatchGetByIds", t, func() {
		ids := []uuid.UUID{}
		for range 10 {
			id := uuid.NewUUID()
			err := testMsgDao.Create(t.Context(), &MsgPO{
				Id:      id,
				Type:    model.MsgText,
				Status:  model.MsgStatusNormal,
				Sender:  1,
				Mtime:   1,
				Content: []byte("hello"),
			})
			So(err, ShouldBeNil)
			ids = append(ids, id)
		}

		msgs, err := testMsgDao.BatchGetByIds(t.Context(), ids)
		So(err, ShouldBeNil)
		So(msgs, ShouldHaveLength, 10)
		for _, msg := range msgs {
			So(msg, ShouldNotBeNil)
		}
	})
}
