package system

import (
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSystemMsgDao_Create(t *testing.T) {
	Convey("TestSystemMsgDao_Create", t, func() {
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		msg := &SystemMsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0, // 系统消息的发送者通常为0
			RecvUid:      10010,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      "测试系统消息",
			Mtime:        time.Now().UnixMicro(),
		}

		err := systemMsgDao.Create(ctx, msg)
		So(err, ShouldBeNil)
	})
}

func TestSystemMsgDao_BatchCreate(t *testing.T) {
	Convey("TestSystemMsgDao_BatchCreate", t, func() {
		systemChatId := uuid.NewUUID()
		msgs := make([]*SystemMsgPO, 3)
		for i := range 3 {
			msgs[i] = &SystemMsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10011,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      "测试批量系统消息" + string(rune('0'+i)),
				Mtime:        time.Now().UnixMicro(),
			}
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		err := systemMsgDao.BatchCreate(ctx, msgs)
		So(err, ShouldBeNil)
	})
}

func TestSystemMsgDao_GetById(t *testing.T) {
	Convey("TestSystemMsgDao_GetById", t, func() {
		// 先创建一条消息
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		content := "用于测试GetById的消息"
		msg := &SystemMsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10012,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      content,
			Mtime:        time.Now().UnixMicro(),
		}

		err := systemMsgDao.Create(ctx, msg)
		So(err, ShouldBeNil)

		// 然后查询
		result, err := systemMsgDao.GetById(ctx, msgId)
		So(err, ShouldBeNil)
		So(result.Id, ShouldEqual, msgId)
		So(result.Content, ShouldEqual, content)
	})
}

func TestSystemMsgDao_BatchGetByIds(t *testing.T) {
	Convey("TestSystemMsgDao_BatchGetByIds", t, func() {
		// 先创建两条消息
		systemChatId := uuid.NewUUID()
		msgIds := make([]uuid.UUID, 2)
		for i := range 2 {
			msgIds[i] = uuid.NewUUID()
			msg := &SystemMsgPO{
				Id:           msgIds[i],
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10013,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      "批量获取消息测试" + string(rune('0'+i)),
				Mtime:        time.Now().UnixMicro(),
			}

			err := systemMsgDao.Create(ctx, msg)
			So(err, ShouldBeNil)
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		// 然后批量查询
		results, err := systemMsgDao.BatchGetByIds(ctx, msgIds)
		So(err, ShouldBeNil)
		So(len(results), ShouldEqual, 2)
	})
}

func TestSystemMsgDao_ListByChatId(t *testing.T) {
	Convey("TestSystemMsgDao_ListByChatId", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10014,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 再创建几条消息
		for i := 0; i < 5; i++ {
			msg := &SystemMsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: chatId,
				Uid:          0,
				RecvUid:      10014,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      "会话消息" + string(rune('0'+i)),
				Mtime:        time.Now().UnixMicro(),
			}

			err = systemMsgDao.Create(ctx, msg)
			So(err, ShouldBeNil)
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		// 然后按会话ID查询
		msgs, err := systemMsgDao.ListByChatId(ctx, chatId, time.Now().UnixMicro(), 10)
		So(err, ShouldBeNil)
		So(len(msgs), ShouldBeGreaterThanOrEqualTo, 5)
	})
}

func TestSystemMsgDao_UpdateStatus(t *testing.T) {
	Convey("TestSystemMsgDao_UpdateStatus", t, func() {
		// 先创建一条消息
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		msg := &SystemMsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10015,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      "需要更新状态的消息",
			Mtime:        time.Now().UnixMicro(),
		}

		err := systemMsgDao.Create(ctx, msg)
		So(err, ShouldBeNil)

		// 然后更新状态
		err = systemMsgDao.UpdateStatus(ctx, msgId, model.SystemMsgStatusRead)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedMsg, err := systemMsgDao.GetById(ctx, msgId)
		So(err, ShouldBeNil)
		So(updatedMsg.Status, ShouldEqual, model.SystemMsgStatusRead)
	})
}

func TestSystemMsgDao_BatchUpdateStatus(t *testing.T) {
	Convey("TestSystemMsgDao_BatchUpdateStatus", t, func() {
		// 先创建几条消息
		systemChatId := uuid.NewUUID()
		msgIds := make([]uuid.UUID, 3)
		for i := 0; i < 3; i++ {
			msgIds[i] = uuid.NewUUID()
			msg := &SystemMsgPO{
				Id:           msgIds[i],
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10016,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      "批量更新状态消息" + string(rune('0'+i)),
				Mtime:        time.Now().UnixMicro(),
			}

			err := systemMsgDao.Create(ctx, msg)
			So(err, ShouldBeNil)
		}

		// 然后批量更新状态
		err := systemMsgDao.BatchUpdateStatus(ctx, msgIds, model.SystemMsgStatusRead)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedMsgs, err := systemMsgDao.BatchGetByIds(ctx, msgIds)
		So(err, ShouldBeNil)
		for _, msg := range updatedMsgs {
			So(msg.Status, ShouldEqual, model.SystemMsgStatusRead)
		}
	})
}

func TestSystemMsgDao_Delete(t *testing.T) {
	Convey("TestSystemMsgDao_Delete", t, func() {
		// 先创建一条消息
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		msg := &SystemMsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10017,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      "需要删除的消息",
			Mtime:        time.Now().UnixMicro(),
		}

		err := systemMsgDao.Create(ctx, msg)
		So(err, ShouldBeNil)

		// 然后删除
		err = systemMsgDao.Delete(ctx, msgId)
		So(err, ShouldBeNil)

		// 验证删除结果
		_, err = systemMsgDao.GetById(ctx, msgId)
		So(err, ShouldNotBeNil)
	})
}

func TestSystemMsgDao_DeleteByChatId(t *testing.T) {
	Convey("TestSystemMsgDao_DeleteByChatId", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &SystemChatPO{
			Id:            chatId,
			Type:          model.SystemNotificationChat,
			Uid:           10018,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			LastReadTime:  time.Now().UnixMicro(),
			UnreadCount:   0,
		}

		err := systemChatDao.Create(ctx, chat)
		So(err, ShouldBeNil)

		// 再创建几条消息
		for i := 0; i < 3; i++ {
			msg := &SystemMsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: chatId,
				Uid:          0,
				RecvUid:      10018,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      "需要按会话删除的消息" + string(rune('0'+i)),
				Mtime:        time.Now().UnixMicro(),
			}

			err = systemMsgDao.Create(ctx, msg)
			So(err, ShouldBeNil)
		}

		// 然后按会话ID删除所有消息
		err = systemMsgDao.DeleteByChatId(ctx, chatId)
		So(err, ShouldBeNil)

		// 验证删除结果
		msgs, err := systemMsgDao.ListByChatId(ctx, chatId, time.Now().UnixMicro(), 10)
		So(err, ShouldBeNil)
		So(len(msgs), ShouldEqual, 0)
	})
}
