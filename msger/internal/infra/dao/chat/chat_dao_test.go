package chat

import (
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChatDao_Create(t *testing.T) {
	Convey("TestChatDao_Create", t, func() {
		err := testChatDao.Create(t.Context(), &ChatPO{
			Id:        uuid.NewUUID(),
			Type:      model.GroupChat,
			Name:      "test",
			Status:    model.ChatStatusNormal,
			Creator:   100,
			Mtime:     time.Now().UnixNano(),
			LastMsgId: uuid.NewUUID(),
			Settings:  0,
		})

		So(err, ShouldBeNil)
	})
}

func TestChatDao_GetById(t *testing.T) {
	Convey("TestChatDao_GetById", t, func() {
		now := time.Now().UnixNano()
		id := uuid.NewUUID()
		lastMsgId := uuid.NewUUID()
		err := testChatDao.Create(t.Context(), &ChatPO{
			Id:        id,
			Type:      model.GroupChat,
			Name:      "test",
			Status:    model.ChatStatusNormal,
			Creator:   100,
			Mtime:     now,
			LastMsgId: lastMsgId,
			Settings:  10,
		})
		So(err, ShouldBeNil)

		chat, err := testChatDao.GetById(t.Context(), id)
		So(err, ShouldBeNil)
		So(chat, ShouldNotBeNil)
		// check chat should equal
		So(chat.Id, ShouldEqual, id)
		So(chat.Type, ShouldEqual, model.GroupChat)
		So(chat.Name, ShouldEqual, "test")
		So(chat.Status, ShouldEqual, model.ChatStatusNormal)
		So(chat.Creator, ShouldEqual, 100)
		So(chat.Mtime, ShouldEqual, now)
		So(chat.LastMsgId, ShouldEqual, lastMsgId)
		So(chat.Settings, ShouldEqual, 10)
	})
}
