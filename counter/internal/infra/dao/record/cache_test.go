package record

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCache(t *testing.T) {
	Convey("TestCache", t, func() {
		var (
			bizCode int32 = 100
			uid     int64 = 1
			oid     int64 = 200
		)

		err := testCache.Add(ctx, bizCode, uid, &CacheRecord{
			Act:   ActDo,
			Oid:   200,
			Mtime: time.Now().Unix(),
		})
		So(err, ShouldBeNil)

		// get
		has, err := testCache.ExistsOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(has, ShouldBeTrue)

		err = testCache.RemoveOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)

		has, err = testCache.ExistsOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(has, ShouldBeFalse)

		testCache.c.Del(getCacheKey(bizCode, uid))
	})
}

func TestBatchExists(t *testing.T) {
	Convey("TestBatchExists", t, func() {

		var (
			bizCode int32   = 100
			uid     int64   = 1
			oids    []int64 = []int64{100, 200, 300, 400, 500, 600}
		)

		defer testCache.c.Del(getCacheKey(bizCode, uid))

		rds := []*CacheRecord{}
		for _, o := range oids {
			rds = append(rds, &CacheRecord{
				Act:   ActDo,
				Mtime: time.Now().Unix(),
				Oid:   o,
			})
		}

		err := testCache.BatchAdd(ctx, bizCode, uid, rds)
		So(err, ShouldBeNil)

		resp, err := testCache.BatchExistsOid(ctx, bizCode, uid, oids...)
		So(err, ShouldBeNil)
		// all should exist
		for _, o := range oids {
			SoMsg(fmt.Sprintf("oid = %d", o), resp[o], ShouldBeTrue)
		}

		// remove some of it
		err = testCache.BatchRemoveOids(ctx, bizCode, uid, 200, 400, 500)
		So(err, ShouldBeNil)

		size, err := testCache.Size(ctx, bizCode, uid)
		So(err, ShouldBeNil)
		So(size, ShouldEqual, 3)

		// batch check exists
		got, err := testCache.BatchExistsOid(ctx, bizCode, uid, oids...) // 200 400 500 should be absent
		So(err, ShouldBeNil)
		cases := map[int64]bool{100: true, 200: false, 300: true, 400: false, 500: false, 600: true}
		for gotOid, gotExist := range got {
			SoMsg(fmt.Sprintf("oid=%d,got=%v", gotOid, gotExist), cases[gotOid], ShouldEqual, gotExist)
		}

	})
}

func TestBatchNotExist(t *testing.T) {
	Convey("TestBatchNotExist", t, func() {
		_, err := testCache.BatchExistsOid(ctx, rand.Int31(), rand.Int63(), 1, 2, 3, 4)
		So(err, ShouldBeNil)
	})
}

func TestSizeLimitBatchAdd(t *testing.T) {
	Convey("TestSizeLimitBatchAdd", t, func() {
		var (
			bizCode int32   = 100
			uid     int64   = 1
			oids    []int64 = []int64{100, 200, 300, 400, 500, 600}
		)
		defer testCache.c.Del(getCacheKey(bizCode, uid))

		rds := []*CacheRecord{}
		for _, o := range oids {
			rds = append(rds, &CacheRecord{
				Act:   ActDo,
				Mtime: time.Now().Unix(),
				Oid:   o,
			})
		}

		testCache.SetMaxMemberPerKey(5)      // 5 members will overflow
		testCache.SetEvitNumberOnOverflow(3) // evit 3 members every time overflow happens

		// batch add first
		err := testCache.BatchAdd(ctx, bizCode, uid, rds)
		So(err, ShouldBeNil)

		size, err := testCache.Size(ctx, bizCode, uid)
		So(err, ShouldBeNil)
		So(size, ShouldEqual, 6)

		newRds := []*CacheRecord{}
		for idx, o := range []int64{700, 800, 900, 1000} { // add 4 more
			newRds = append(newRds, &CacheRecord{
				Act:   ActDo,
				Mtime: time.Now().Unix() + 100 + int64(idx),
				Oid:   o,
			})
		}

		// batch add with size limit
		err = testCache.SizeLimitBatchAdd(ctx, bizCode, uid, newRds)
		So(err, ShouldBeNil)

		// check size
		size, err = testCache.Size(ctx, bizCode, uid)
		So(err, ShouldBeNil)
		So(size, ShouldEqual, 7) // 6 - 3 + 4 = 7

	})
}
