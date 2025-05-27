package p2p

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)


func TestChatDao_Create(t *testing.T) {
	Convey("TestChatDao_Create", t, func() {
		id, err := chatDao.Create(ctx, &Chat{
			ChatId:      100,
			UserId:      10,
			PeerId:      20,
			UnReadCount: 100,
		})
		So(err, ShouldBeNil)
		So(id, ShouldNotBeZeroValue)
	})
}

func TestChatDao_InitChat(t *testing.T) {
	Convey("TestChatDao_InitChat", t, func() {
		err := chatDao.InitChat(ctx, 900, 1000, 300)
		So(err, ShouldBeNil)
	})
}
