package system

import (
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	. "github.com/smartystreets/goconvey/convey"
)

func deleteForTest() {
	systemMsgDao.db.Exec("DELETE FROM system_msg WHERE 1=1")
	systemChatDao.db.Exec("DELETE FROM system_chat WHERE 1=1")
}

func TestSystemChatDao_Create(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_Create", t, func() {
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10001,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)
	})
}

func TestSystemChatDao_GetByUidAndType(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_GetByUidAndType", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10002,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 然后查询
		got, err := systemChatDao.GetByUidAndType(ctx, 10002, model.SystemNotificationChat)
		So(err, ShouldBeNil)
		So(got.Uid, ShouldEqual, 10002)
		So(got.Type, ShouldEqual, model.SystemNotificationChat)
		So(got.Id.String(), ShouldEqual, chatId.String())
	})
}

func TestSystemChatDao_ListByUid(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_ListByUid", t, func() {
		// 先创建两个会话
		uid := int64(10003)
		for i := range 2 {
			chatId := uuid.NewUUID()
			chat := &SystemChatPO{
				Id:            chatId,
				Type:          model.SystemNotificationChat + model.SystemChatType(i),
				Uid:           uid,
				Mtime:         time.Now().UnixMicro(),
				LastMsgId:     uuid.NewUUID(),
				LastReadMsgId: uuid.NewUUID(),
				LastReadTime:  time.Now().UnixMicro(),
				UnreadCount:   int64(i),
			}

			err := systemChatDao.Create(ctx, chat)
			So(err, ShouldBeNil)
			time.Sleep(10 * time.Millisecond) // 确保mtime有差异
		}

		// 然后查询
		results, err := systemChatDao.ListByUid(ctx, uid)
		So(err, ShouldBeNil)
		So(len(results), ShouldBeGreaterThanOrEqualTo, 2)
	})
}

func TestSystemChatDao_UpdateLastMsg(t *testing.T) {
	Convey("TestSystemChatDao_UpdateLastMsg", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10004,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 然后更新
		newLastMsgId := uuid.NewUUID()
		err = systemChatDao.UpdateLastMsg(ctx, chatId, newLastMsgId, true)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedChat, err := systemChatDao.GetByUidAndType(ctx, 10004,
			model.SystemNotificationChat)
		So(err, ShouldBeNil)
		So(updatedChat.LastMsgId.String(), ShouldEqual, newLastMsgId.String())
		So(updatedChat.UnreadCount, ShouldEqual, 1)
	})
}

func TestSystemChatDao_ClearUnread(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_ClearUnread", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10005,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   5,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 然后清空未读
		lastReadMsgId := uuid.NewUUID()
		err = systemChatDao.ClearUnread(ctx, chatId, lastReadMsgId)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedChat, err := systemChatDao.GetByUidAndType(ctx, 10005, model.SystemNotificationChat)
		So(err, ShouldBeNil)
		So(updatedChat.UnreadCount, ShouldEqual, 0)
		So(updatedChat.LastReadMsgId, ShouldEqual, lastReadMsgId)
	})
}

func TestSystemChatDao_Delete(t *testing.T) {
	Convey("TestSystemChatDao_Delete", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10006,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 然后删除
		err = systemChatDao.Delete(ctx, chatId)
		So(err, ShouldBeNil)

		// 验证删除结果
		_, err = systemChatDao.GetByUidAndType(ctx, 10006, model.SystemNotificationChat)
		So(err, ShouldNotBeNil)
	})
}
