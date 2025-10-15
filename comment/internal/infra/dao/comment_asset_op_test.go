package dao

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAssetDao(t *testing.T) {
	Convey("TestAssetDao", t, func() {
		err := testCommentAssetDao.BatchInsert(testCtx, []*CommentAsset{
			{Type: 1, StoreKey: "abc", CommentId: 100, Metadata: json.RawMessage{}},
			{Type: 1, StoreKey: "efg", CommentId: 100, Metadata: json.RawMessage{}},
			{Type: 1, StoreKey: "hij", CommentId: 100, Metadata: json.RawMessage{}},
		})
		So(err, ShouldBeNil)

		gots, err := testCommentAssetDao.GetByCommentId(testCtx, 100)
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 3)
		for _, g := range gots {
			t.Log(g)
		}

		err = testCommentAssetDao.DeleteByCommentId(testCtx, 100)
		So(err, ShouldBeNil)

		gots, err = testCommentAssetDao.GetByCommentId(testCtx, 100)
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 0)

		err = testCommentAssetDao.BatchInsert(testCtx, []*CommentAsset{
			{Type: 1, StoreKey: "abc", CommentId: 100, Metadata: json.RawMessage{}},
			{Type: 1, StoreKey: "efg", CommentId: 200, Metadata: json.RawMessage{}},
			{Type: 1, StoreKey: "hij", CommentId: 300, Metadata: json.RawMessage{}},
		})
		So(err, ShouldBeNil)

		err = testCommentAssetDao.BatchDeleteByCommentId(testCtx, []int64{100, 200, 300})
		So(err, ShouldBeNil)

		for _, id := range []int64{100, 200, 300} {
			gots, err = testCommentAssetDao.GetByCommentId(testCtx, id)
			So(err, ShouldBeNil)
			So(len(gots), ShouldEqual, 0)
		}
	})
}
