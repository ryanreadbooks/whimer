package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRelationSettingDao_GetByUid(t *testing.T) {
	Convey("GetByUid", t, func() {
		err := testSettingDao.Insert(ctx, &RelationSetting{
			Uid:      100,
			Settings: DefaultSettings.Json(),
		})
		So(err, ShouldBeNil)
		got, e := testSettingDao.Get(ctx, 100)
		So(e, ShouldBeNil)
		So(got.ParseSettings().DisplayFanList, ShouldEqual, DefaultSettings.DisplayFanList)
		So(got.ParseSettings().DisplayFollowingList, ShouldEqual, DefaultSettings.DisplayFollowingList)

		testSettingDao.Delete(ctx, 100)
	})
}

func TestRelationSettingDao_Insert(t *testing.T) {
	Convey("Insert", t, func() {
		err := testSettingDao.Insert(ctx, &RelationSetting{
			Uid:      100,
			Settings: DefaultSettings.Json(),
		})
		So(err, ShouldBeNil)
		testSettingDao.Delete(ctx, 100)
	})
}

func TestRelationSettingDao_Update(t *testing.T) {
	Convey("Update", t, func() {
		err := testSettingDao.Insert(ctx, &RelationSetting{
			Uid:      100,
			Settings: DefaultSettings.Json(),
		})
		So(err, ShouldBeNil)

		err = testSettingDao.Update(ctx, &RelationSetting{
			Uid: 100,
			Settings: (&Settings{
				DisplayFanList:       true,
				DisplayFollowingList: false,
			}).Json(),
		})
		So(err, ShouldBeNil)

		got, err := testSettingDao.Get(ctx, 100)
		So(err, ShouldBeNil)
		setting := got.ParseSettings()
		So(setting.DisplayFollowingList, ShouldBeFalse)

		testSettingDao.Delete(ctx, 100)
	})
}
