package note

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNoteAssetRepo_BatchUpdateAssetMeta(t *testing.T) {
	Convey("NoteAssetRepo BatchUpdateAssetMeta", t, func() {
		// 插入三条
		err := noteAssetRepo.BatchInsert(ctx, []*AssetPO{
			{
				AssetKey:  "1",
				NoteId:    100,
				AssetMeta: []byte("1"),
			},
			{
				AssetKey:  "2",
				NoteId:    100,
				AssetMeta: []byte("2"),
			},
			{
				AssetKey:  "3",
				NoteId:    100,
				AssetMeta: []byte("3"),
			},
		})
		So(err, ShouldBeNil)

		//  批量更新
		err = noteAssetRepo.BatchUpdateAssetMeta(ctx, 100, map[string][]byte{
			"1": []byte("100"),
			"2": []byte("200"),
			"3": []byte("300"),
		})
		So(err, ShouldBeNil)

		// 查询
		assets, err := noteAssetRepo.FindByNoteIds(ctx, []int64{100})
		So(err, ShouldBeNil)
		So(len(assets), ShouldEqual, 3)
		// 检查每个的asset_meta
		for _, asset := range assets {
			switch asset.AssetKey {
			case "1":
				So(string(asset.AssetMeta), ShouldEqual, "100")
			case "2":
				So(string(asset.AssetMeta), ShouldEqual, "200")
			case "3":
				So(string(asset.AssetMeta), ShouldEqual, "300")
			}
		}
	})
}
