package dao

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/ryanreadbooks/whimer/comment/internal/model"
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

func TestAssetCache(t *testing.T) {
	a := CommentAsset{
		Id:        100,
		CommentId: 100,
		Type:      model.CommentAssetCustomEmoji,
		StoreKey:  "abc",
		Metadata:  json.RawMessage{},
		Ctime:     100,
	}

	err := testCache.Hmset("test_asset_dao_key_1", map[string]string{
		"id":         strconv.FormatInt(a.Id, 10),
		"comment_id": strconv.FormatInt(a.CommentId, 10),
		"type":       strconv.FormatInt(int64(a.Type), 10),
		"store_key":  a.StoreKey,
		"metadata":   string(a.Metadata),
		"ctime":      strconv.FormatInt(a.Ctime, 10),
	})

	if err != nil {
		t.Fatal(err)
	}

	pipe, err := testCache.TxPipeline()
	if err != nil {
		t.Fatal(err)
	}

	res := pipe.HGetAll(testCtx, "test_asset_dao_key_1")
	_, err = pipe.Exec(testCtx)
	if err != nil {
		t.Fatal(err)
	}

	// scan
	var got CommentAsset
	err = res.Scan(&got)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", got)

}
