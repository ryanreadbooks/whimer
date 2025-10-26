package system

import (
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func deleteForTest() {
	testSystemMsgDao.db.Exec("DELETE FROM system_msg WHERE 1=1")
	testSystemChatDao.db.Exec("DELETE FROM system_chat WHERE 1=1")
}

func TestSystemChatDao_Create(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_Create", t, func() {
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10001,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)
	})
}

func TestSystemChatDao_GetByUidAndType(t *testing.T) {
	defer deleteForTest()
	Convey("TestSystemChatDao_GetByUidAndType", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10002,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 然后查询
		got, err := testSystemChatDao.GetByUidAndType(textctx, 10002, model.SystemNotifyNoticeChat)
		So(err, ShouldBeNil)
		So(got.Uid, ShouldEqual, 10002)
		So(got.Type, ShouldEqual, model.SystemNotifyNoticeChat)
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
			chat := &ChatPO{
				Id:            chatId,
				Type:          model.SystemNotifyNoticeChat + model.SystemChatType(i),
				Uid:           uid,
				Mtime:         time.Now().UnixMicro(),
				LastMsgId:     uuid.NewUUID(),
				LastReadMsgId: uuid.NewUUID(),
				UnreadCount:   int64(i),
			}

			err := testSystemChatDao.Create(textctx, chat)
			So(err, ShouldBeNil)
			time.Sleep(10 * time.Millisecond) // 确保mtime有差异
		}

		// 然后查询
		results, err := testSystemChatDao.ListByUid(textctx, uid)
		So(err, ShouldBeNil)
		So(len(results), ShouldBeGreaterThanOrEqualTo, 2)
	})
}

func TestSystemChatDao_UpdateLastMsg(t *testing.T) {
	Convey("TestSystemChatDao_UpdateLastMsg", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10004,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 然后更新
		newLastMsgId := uuid.NewUUID()
		err = testSystemChatDao.UpdateLastMsg(textctx, chatId, newLastMsgId, true)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedChat, err := testSystemChatDao.GetByUidAndType(textctx, 10004,
			model.SystemNotifyNoticeChat)
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
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10005,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   5,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 然后清空未读
		lastReadMsgId := uuid.NewUUID()
		err = testSystemChatDao.ClearUnread(textctx, chatId, lastReadMsgId)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedChat, err := testSystemChatDao.GetByUidAndType(textctx, 10005, model.SystemNotifyNoticeChat)
		So(err, ShouldBeNil)
		So(updatedChat.UnreadCount, ShouldEqual, 0)
		So(updatedChat.LastReadMsgId, ShouldEqual, lastReadMsgId)
	})
}

func TestSystemChatDao_Delete(t *testing.T) {
	Convey("TestSystemChatDao_Delete", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10006,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 然后删除
		err = testSystemChatDao.Delete(textctx, chatId)
		So(err, ShouldBeNil)

		// 验证删除结果
		_, err = testSystemChatDao.GetByUidAndType(textctx, 10006, model.SystemNotifyNoticeChat)
		So(err, ShouldNotBeNil)
	})
}
