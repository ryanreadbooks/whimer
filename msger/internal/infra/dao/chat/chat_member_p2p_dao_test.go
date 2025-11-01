package chat

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChatMemberP2PDao_Create(t *testing.T) {
	Convey("TestChatMemberP2PDao_Create", t, func() {
		chatId := uuid.NewUUID()
		member := &ChatMemberP2PPO{
			ChatId: chatId,
			UidA:   1,
			UidB:   2,
			Ctime:  time.Now().Unix(),
			Mtime:  time.Now().Unix(),
		}
		err := testChatMemberP2PDao.Create(t.Context(), member)
		So(err, ShouldBeNil)
	})
}

func TestChatMemberP2PDao_Get(t *testing.T) {
	Convey("TestChatMemberP2PDao_Get", t, func() {
		chatId := uuid.NewUUID()
		uidA := rand.Int63()
		uidB := rand.Int63()
		member := &ChatMemberP2PPO{
			ChatId: chatId,
			UidA:   uidA,
			UidB:   uidB,
			Ctime:  time.Now().Unix(),
			Mtime:  time.Now().Unix(),
		}
		err := testChatMemberP2PDao.Create(t.Context(), member)
		So(err, ShouldBeNil)

		got, err := testChatMemberP2PDao.GetByChatId(t.Context(), chatId)
		So(err, ShouldBeNil)
		So(member.ChatId, ShouldResemble, got.ChatId)

		// by uids
		got, err = testChatMemberP2PDao.GetByUids(t.Context(), uidA, uidB)
		So(err, ShouldBeNil)
		So(member.ChatId, ShouldEqual, got.ChatId)
	})
}

func TestChatMemberP2PDao_GetByChatIdUid(t *testing.T) {
	Convey("TestChatMemberP2PDao_GetByChatIdUid", t, func() {
		chatId := uuid.NewUUID()
		uidA := rand.Int63()
		uidB := rand.Int63()
		member := &ChatMemberP2PPO{
			ChatId: chatId,
			UidA:   uidA,
			UidB:   uidB,
			Ctime:  time.Now().Unix(),
			Mtime:  time.Now().Unix(),
		}
		err := testChatMemberP2PDao.Create(t.Context(), member)
		So(err, ShouldBeNil)

		got, err := testChatMemberP2PDao.GetByChatIdUid(t.Context(), chatId, uidA)
		So(err, ShouldBeNil)
		got2, err := testChatMemberP2PDao.GetByChatIdUid(t.Context(), chatId, uidB)
		So(err, ShouldBeNil)

		So(got.ChatId, ShouldEqual, got2.ChatId)
		So(got.UidA, ShouldEqual, got2.UidA)
		So(got.UidB, ShouldEqual, got2.UidB)
	})
}
