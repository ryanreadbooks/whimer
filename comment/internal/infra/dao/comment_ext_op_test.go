package dao

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmihailenco/msgpack/v5"
)

func TestJsonMarshalExt(t *testing.T) {
	ext := &CommentExt{
		CommentId: 109012,
		AtUsers:   json.RawMessage(`{"name":"age","age":12,"address":"earth"}`),
	}

	b, e := json.Marshal(ext)
	t.Log(e)
	t.Log(string(b), len(b))

	b, e = msgpack.Marshal(ext)
	t.Log(e)
	t.Log(string(b), len(b))
}

func TestExtDao(t *testing.T) {
	data := json.RawMessage(`{"name":"age"}`)
	dataStr := string(data)
	Convey("TestExtDao", t, func() {
		err := testCommentExtDao.Upsert(testCtx, &CommentExt{
			CommentId: 1,
			AtUsers:   data,
		})
		So(err, ShouldBeNil)

		got, err := testCommentExtDao.Get(testCtx, 1)
		So(err, ShouldBeNil)
		So(string(got.AtUsers), ShouldEqual, dataStr)

		err = testCommentExtDao.Delete(testCtx, 1)
		So(err, ShouldBeNil)

		//
		err = testCommentExtDao.Upsert(testCtx, &CommentExt{
			CommentId: 3,
			AtUsers:   data,
		})
		So(err, ShouldBeNil)
		err = testCommentExtDao.Upsert(testCtx, &CommentExt{
			CommentId: 4,
			AtUsers:   data,
		})
		So(err, ShouldBeNil)

		gots, err := testCommentExtDao.BatchGet(testCtx, []int64{3, 4})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 2)
		So(string(gots[0].AtUsers), ShouldEqual, dataStr)
		So(string(gots[1].AtUsers), ShouldEqual, dataStr)
	})
}
