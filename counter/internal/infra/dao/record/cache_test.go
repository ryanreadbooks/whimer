package record

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCache(t *testing.T) {
	Convey("TestCache", t, func() {
		var (
			bizCode int32 = 100
			uid     int64 = 1
			oid     int64 = 200
		)

		err := testCache.CounterListAdd(ctx, bizCode, uid, &CacheRecord{
			Act:   ActDo,
			Oid:   200,
			Mtime: time.Now().Unix(),
		})
		So(err, ShouldBeNil)

		// get
		has, err := testCache.CounterListExistsOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(has, ShouldBeTrue)

		err = testCache.CounterListRemoveOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)

		has, err = testCache.CounterListExistsOid(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(has, ShouldBeFalse)

		testCache.c.Del(testCache.getCounterListCacheKey(bizCode, uid))
	})
}

func TestBatchExists(t *testing.T) {
	Convey("TestBatchExists", t, func() {

		var (
			bizCode int32   = 100
			uid     int64   = 1
			oids    []int64 = []int64{100, 200, 300, 400, 500, 600}
		)

		defer testCache.c.Del(testCache.getCounterListCacheKey(bizCode, uid))

		rds := []*CacheRecord{}
		for _, o := range oids {
			rds = append(rds, &CacheRecord{
				Act:   ActDo,
				Mtime: time.Now().Unix(),
				Oid:   o,
			})
		}

		err := testCache.CounterListBatchAdd(ctx, bizCode, uid, rds)
		So(err, ShouldBeNil)

		resp, err := testCache.CounterListBatchExistsOid(ctx, bizCode, uid, oids...)
		So(err, ShouldBeNil)
		// all should exist
		for _, o := range oids {
			SoMsg(fmt.Sprintf("oid = %d", o), resp[o], ShouldBeTrue)
		}

		// remove some of it
		err = testCache.CounterListBatchRemoveOids(ctx, bizCode, uid, 200, 400, 500)
		So(err, ShouldBeNil)

		size, err := testCache.CounterListSize(ctx, bizCode, uid)
		So(err, ShouldBeNil)
		So(size, ShouldEqual, 3)

		// batch check exists
		got, err := testCache.CounterListBatchExistsOid(ctx, bizCode, uid, oids...) // 200 400 500 should be absent
		So(err, ShouldBeNil)
		cases := map[int64]bool{100: true, 200: false, 300: true, 400: false, 500: false, 600: true}
		for gotOid, gotExist := range got {
			SoMsg(fmt.Sprintf("oid=%d,got=%v", gotOid, gotExist), cases[gotOid], ShouldEqual, gotExist)
		}

	})
}

func TestBatchNotExist(t *testing.T) {
	Convey("TestBatchNotExist", t, func() {
		_, err := testCache.CounterListBatchExistsOid(ctx, rand.Int31(), rand.Int63(), 1, 2, 3, 4)
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
		defer testCache.c.Del(testCache.getCounterListCacheKey(bizCode, uid))

		rds := []*CacheRecord{}
		for _, o := range oids {
			rds = append(rds, &CacheRecord{
				Act:   ActDo,
				Mtime: time.Now().Unix(),
				Oid:   o,
			})
		}

		testCache.SetCounterListMaxMember(5)  // 5 members will overflow
		testCache.SetCounterListEvitNumber(3) // evit 3 members every time overflow happens

		// batch add first
		err := testCache.CounterListBatchAdd(ctx, bizCode, uid, rds)
		So(err, ShouldBeNil)

		size, err := testCache.CounterListSize(ctx, bizCode, uid)
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
		err = testCache.CounterListSizeLimitBatchAdd(ctx, bizCode, uid, newRds)
		So(err, ShouldBeNil)

		// check size
		size, err = testCache.CounterListSize(ctx, bizCode, uid)
		So(err, ShouldBeNil)
		So(size, ShouldEqual, 7) // 6 - 3 + 4 = 7

	})
}

func TestAddRecord(t *testing.T) {
	Convey("TestAddRecord", t, func() {
		var (
			bizCode int32 = 2000
			uid     int64 = 100
			oid     int64 = 200
		)
		err := testCache.AddRecord(ctx, &Record{
			Id:      100,
			Act:     ActDo,
			Uid:     uid,
			Oid:     oid,
			BizCode: bizCode,
			Mtime:   time.Now().Unix(),
			Ctime:   time.Now().Unix(),
		})
		So(err, ShouldBeNil)

		record, err := testCache.GetRecord(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(record.Id, ShouldEqual, 100)
		t.Log(record)
	})
}

func TestGetRecord(t *testing.T) {
	Convey("TestGetRecord", t, func() {
		var (
			bizCode int32 = 2000
			uid     int64 = 100
			oid     int64 = 200
		)
		record, err := testCache.GetRecord(ctx, bizCode, uid, oid)
		So(errors.Is(err, redis.Nil), ShouldBeTrue)
		So(record, ShouldBeNil)
	})
}

func TestBatchAddRecord(t *testing.T) {
	Convey("TestBatchAddRecord\n", t, func() {
		var (
			bizCode int32 = 2000
		)

		var records = []*Record{}
		records = append(records, &Record{
			Id:      1,
			BizCode: bizCode,
			Act:     ActDo,
			Uid:     100,
			Oid:     200,
			Ctime:   time.Now().Unix(),
			Mtime:   time.Now().Unix(),
		})
		records = append(records, &Record{
			Id:      2,
			BizCode: bizCode,
			Uid:     200,
			Oid:     200,
			Act:     ActUndo,
			Ctime:   time.Now().Unix(),
			Mtime:   time.Now().Unix(),
		})
		err := testCache.BatchAddRecord(ctx, records)
		So(err, ShouldBeNil)

		key1 := CacheKey{BizCode: bizCode, Uid: 100, Oid: 200}
		key2 := CacheKey{BizCode: bizCode, Uid: 200, Oid: 200}
		gots, err := testCache.BatchGetRecord(ctx, []CacheKey{key1, key2})
		So(err, ShouldBeNil)
		for k, v := range gots {
			t.Logf("k=%v, v=%+v\n", k, v)
		}
		So(len(gots), ShouldEqual, 2)
		So(gots[key1].Act, ShouldEqual, ActDo)
		So(gots[key2].Act, ShouldEqual, ActUndo)
	})
}

func TestBatchGetRecord(t *testing.T) {
	Convey("TestBatchGetRecord", t, func() {
		var (
			bizCode int32 = 2000
			uid     int64 = 100
			oid     int64 = 200
		)
		err := testCache.AddRecord(ctx, &Record{
			Act:     ActUndo,
			Uid:     uid,
			Oid:     oid,
			BizCode: bizCode,
			Mtime:   time.Now().Unix(),
			Ctime:   time.Now().Unix(),
		})
		So(err, ShouldBeNil)

		gots, err := testCache.BatchGetRecord(ctx, []CacheKey{
			{BizCode: bizCode, Uid: uid, Oid: oid},
			{BizCode: 10086, Uid: 10086, Oid: 10086}})

		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 1)
	})
}

func TestCheckHasCountedRecord(t *testing.T) {
	Convey("TestCheckHasCountedRecord", t, func() {
		var (
			bizCode int32 = 100
			uid     int64 = 999
			oid     int64 = 888
		)
		counterListKey := testCache.getCounterListCacheKey(bizCode, uid)
		counterRecordKey := testCache.getCounterRecordCacheKey(bizCode, uid, oid)
		t.Log(counterRecordKey)
		t.Log(counterListKey)
		defer testCache.c.Del(counterListKey, counterRecordKey)

		err := testCache.AddRecord(ctx, &Record{
			BizCode: bizCode,
			Uid:     uid,
			Oid:     oid,
			Act:     ActDo,
			Ctime:   time.Now().Unix(),
			Mtime:   time.Now().Unix(),
		})
		So(err, ShouldBeNil)

		got, err := testCache.CheckHasCountedRecord(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(got, ShouldBeTrue)
	})

	Convey("TestCheckHasCountedRecord 2", t, func() {
		var (
			bizCode int32 = 100
			uid     int64 = 999
			oid     int64 = 888
		)
		counterListKey := testCache.getCounterListCacheKey(bizCode, uid)
		counterRecordKey := testCache.getCounterRecordCacheKey(bizCode, uid, oid)
		t.Log(counterRecordKey)
		t.Log(counterListKey)
		defer testCache.c.Del(counterListKey, counterRecordKey)

		err := testCache.CounterListAdd(ctx, bizCode, uid, &CacheRecord{Act: ActDo, Oid: oid, Mtime: time.Now().Unix()})
		So(err, ShouldBeNil)

		got, err := testCache.CheckHasCountedRecord(ctx, bizCode, uid, oid)
		So(err, ShouldBeNil)
		So(got, ShouldBeTrue)
	})
}
