package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRelationSettingDao_GetByUid(t *testing.T) {
	Convey("GetByUid", t, func() {
		_, e := settingDao.Get(ctx, 100)
		So(e, ShouldBeNil)
	})
}

func TestRelationSettingDao_Insert(t *testing.T) {
	Convey("Insert", t, func() {
		err := settingDao.Insert(ctx, &RelationSetting{
			Uid:               100,
			NotShowFollowings: NotShowFollowings,
			ShowFans:          NotShowFans,
		})
		So(err, ShouldBeNil)
	})
}

func TestRelationSettingDao_Update(t *testing.T) {
	Convey("Update", t, func() {
		err := settingDao.Update(ctx, &RelationSetting{
			Uid:               100,
			NotShowFollowings: ShowFollowings,
			ShowFans:          ShowFans,
		})
		So(err, ShouldBeNil)
	})
}
