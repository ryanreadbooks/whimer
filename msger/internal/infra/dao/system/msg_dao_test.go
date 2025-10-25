package system

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSystemMsgDao_Create(t *testing.T) {
	Convey("TestSystemMsgDao_Create", t, func() {
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		msg := &MsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0, // 系统消息的发送者通常为0
			RecvUid:      10010,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      json.RawMessage(`"测试系统消息"`),
			Mtime:        time.Now().UnixMicro(),
		}

		err := testSystemMsgDao.Create(textctx, msg)
		So(err, ShouldBeNil)
	})
}

func TestSystemMsgDao_BatchCreate(t *testing.T) {
	Convey("TestSystemMsgDao_BatchCreate", t, func() {
		systemChatId := uuid.NewUUID()
		msgs := make([]*MsgPO, 3)
		for i := range 3 {
			msgs[i] = &MsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10011,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      json.RawMessage(`"测试批量系统消息"` + string(rune('0'+i))),
				Mtime:        time.Now().UnixMicro(),
			}
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		err := testSystemMsgDao.BatchCreate(textctx, msgs)
		So(err, ShouldBeNil)
	})
}

func TestSystemMsgDao_GetById(t *testing.T) {
	Convey("TestSystemMsgDao_GetById", t, func() {
		// 先创建一条消息
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		content := "用于测试GetById的消息"
		msg := &MsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10012,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      json.RawMessage(content),
			Mtime:        time.Now().UnixMicro(),
		}

		err := testSystemMsgDao.Create(textctx, msg)
		So(err, ShouldBeNil)

		// 然后查询
		result, err := testSystemMsgDao.GetById(textctx, msgId)
		So(err, ShouldBeNil)
		So(result.Id, ShouldEqual, msgId)
		So(string(result.Content), ShouldEqual, string(content))

		chatId, err := testSystemMsgDao.GetChatIdById(textctx, msgId)
		So(err, ShouldBeNil)
		So(chatId.EqualsTo(systemChatId), ShouldBeTrue)
		t.Log(chatId.String())
	})
}

func TestSystemMsgDao_BatchGetByIds(t *testing.T) {
	Convey("TestSystemMsgDao_BatchGetByIds", t, func() {
		// 先创建两条消息
		systemChatId := uuid.NewUUID()
		msgIds := make([]uuid.UUID, 2)
		for i := range 2 {
			msgIds[i] = uuid.NewUUID()
			msg := &MsgPO{
				Id:           msgIds[i],
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10013,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      json.RawMessage(`"批量获取消息测试"` + string(rune('0'+i))),
				Mtime:        time.Now().UnixMicro(),
			}

			err := testSystemMsgDao.Create(textctx, msg)
			So(err, ShouldBeNil)
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		// 然后批量查询
		results, err := testSystemMsgDao.BatchGetByIds(textctx, msgIds)
		So(err, ShouldBeNil)
		So(len(results), ShouldEqual, 2)
	})
}

func TestSystemMsgDao_ListByChatId(t *testing.T) {
	Convey("TestSystemMsgDao_ListByChatId", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10014,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 再创建几条消息
		for i := 0; i < 5; i++ {
			msg := &MsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: chatId,
				Uid:          0,
				RecvUid:      10014,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      json.RawMessage(`"会话消息"` + string(rune('0'+i))),
				Mtime:        time.Now().UnixMicro(),
			}

			err = testSystemMsgDao.Create(textctx, msg)
			So(err, ShouldBeNil)
			time.Sleep(1 * time.Millisecond) // 确保mtime有差异
		}

		// 然后按会话ID查询
		msgs, err := testSystemMsgDao.ListByChatId(textctx, chatId, uuid.MaxUUID(), 10)
		So(err, ShouldBeNil)
		So(len(msgs), ShouldBeGreaterThanOrEqualTo, 5)
	})
}

func TestSystemMsgDao_UpdateStatus(t *testing.T) {
	Convey("TestSystemMsgDao_UpdateStatus", t, func() {
		// 先创建一条消息
		msgId := uuid.NewUUID()
		systemChatId := uuid.NewUUID()
		msg := &MsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10015,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      json.RawMessage(`"需要更新状态的消息"`),
			Mtime:        time.Now().UnixMicro(),
		}

		err := testSystemMsgDao.Create(textctx, msg)
		So(err, ShouldBeNil)

		// 然后更新状态
		err = testSystemMsgDao.UpdateStatus(textctx, msgId, model.SystemMsgStatusRead)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedMsg, err := testSystemMsgDao.GetById(textctx, msgId)
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
			msg := &MsgPO{
				Id:           msgIds[i],
				SystemChatId: systemChatId,
				Uid:          0,
				RecvUid:      10016,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      json.RawMessage(`"批量更新状态消息"` + string(rune('0'+i))),
				Mtime:        time.Now().UnixMicro(),
			}

			err := testSystemMsgDao.Create(textctx, msg)
			So(err, ShouldBeNil)
		}

		// 然后批量更新状态
		err := testSystemMsgDao.BatchUpdateStatus(textctx, msgIds, model.SystemMsgStatusRead)
		So(err, ShouldBeNil)

		// 验证更新结果
		updatedMsgs, err := testSystemMsgDao.BatchGetByIds(textctx, msgIds)
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
		msg := &MsgPO{
			Id:           msgId,
			SystemChatId: systemChatId,
			Uid:          0,
			RecvUid:      10017,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      model.MsgText,
			Content:      json.RawMessage(`"需要删除的消息"`),
			Mtime:        time.Now().UnixMicro(),
		}

		err := testSystemMsgDao.Create(textctx, msg)
		So(err, ShouldBeNil)

		// 然后删除
		err = testSystemMsgDao.DeleteById(textctx, msgId)
		So(err, ShouldBeNil)

		// 验证删除结果
		_, err = testSystemMsgDao.GetById(textctx, msgId)
		So(err, ShouldNotBeNil)
	})
}

func TestSystemMsgDao_DeleteByChatId(t *testing.T) {
	Convey("TestSystemMsgDao_DeleteByChatId", t, func() {
		// 先创建一个会话
		chatId := uuid.NewUUID()
		chat := &ChatPO{
			Id:            chatId,
			Type:          model.SystemNotifyNoticeChat,
			Uid:           10018,
			Mtime:         time.Now().UnixMicro(),
			LastMsgId:     uuid.NewUUID(),
			LastReadMsgId: uuid.NewUUID(),
			UnreadCount:   0,
		}

		err := testSystemChatDao.Create(textctx, chat)
		So(err, ShouldBeNil)

		// 再创建几条消息
		for i := 0; i < 3; i++ {
			msg := &MsgPO{
				Id:           uuid.NewUUID(),
				SystemChatId: chatId,
				Uid:          0,
				RecvUid:      10018,
				Status:       model.SystemMsgStatusNormal,
				MsgType:      model.MsgText,
				Content:      json.RawMessage(`"需要按会话删除的消息"` + string(rune('0'+i))),
				Mtime:        time.Now().UnixMicro(),
			}

			err = testSystemMsgDao.Create(textctx, msg)
			So(err, ShouldBeNil)
		}

		// 然后按会话ID删除所有消息
		err = testSystemMsgDao.DeleteByChatId(textctx, chatId)
		So(err, ShouldBeNil)

		// 验证删除结果
		msgs, err := testSystemMsgDao.ListByChatId(textctx, chatId, uuid.EmptyUUID(), 10)
		So(err, ShouldBeNil)
		So(len(msgs), ShouldEqual, 0)
	})
}
