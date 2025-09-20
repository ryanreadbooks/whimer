package dao

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xslice"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRelationDao_Insert(t *testing.T) {
	testRelations := []*Relation{
		newRelationFromAlphaToBeta(1000, 1002),
		newRelationFromAlphaToBeta(1000, 1003),
		newRelationFromBetaToAlpha(1000, 1004),
		newRelationFromAlphaToBeta(1004, 1002),
		newRelationFromAlphaToBeta(1004, 1005),
		newRelationFromAlphaToBeta(1005, 1002),
		newMutualRelation(1005, 1003),
	}
	Convey("Insert", t, func() {
		for _, c := range testRelations {
			err := testRelationDao.Insert(ctx, c)
			So(err, ShouldBeNil)
		}
	})
}

func TestRelationDao_UpdateLink(t *testing.T) {
	testRelations := []*Relation{
		newRelationFromAlphaToBeta(1000, 1002),
		newRelationFromAlphaToBeta(1000, 1003),
		newRelationFromBetaToAlpha(1000, 1004),
		newRelationFromAlphaToBeta(1004, 1002),
		newRelationFromAlphaToBeta(1004, 1005),
		newRelationFromAlphaToBeta(1005, 1002),
		newMutualRelation(1005, 1003),
	}
	for _, c := range testRelations {
		testRelationDao.Insert(ctx, c)
	}
	Convey("UpdateLink", t, func() {
		err := testRelationDao.UpdateLink(ctx, newRelationFromAlphaToBeta(1000, 1004))
		So(err, ShouldBeNil)
	})
}

func TestRelation_FindByAlphaBeta(t *testing.T) {
	testRelations := []*Relation{
		newRelationFromAlphaToBeta(1000, 1002),
		newRelationFromAlphaToBeta(1000, 1003),
		newRelationFromBetaToAlpha(1000, 1004),
		newRelationFromAlphaToBeta(1004, 1002),
		newRelationFromAlphaToBeta(1004, 1005),
		newRelationFromAlphaToBeta(1005, 1002),
		newMutualRelation(1005, 1003),
	}
	for _, c := range testRelations {
		testRelationDao.Insert(ctx, c)
	}
	cases := []struct {
		alpha   int64
		beta    int64
		desire  *Relation
		wantErr bool
	}{
		{
			alpha:  1000,
			beta:   1002,
			desire: &Relation{UserAlpha: 1000, UserBeta: 1002, Link: LinkForward},
		},
		{
			alpha:  1002,
			beta:   1000,
			desire: &Relation{UserAlpha: 1000, UserBeta: 1002, Link: LinkForward},
		},
		{
			alpha:  1005,
			beta:   1003,
			desire: &Relation{UserAlpha: 1003, UserBeta: 1005, Link: LinkMutual},
		},
		{
			alpha:   1005,
			beta:    1000,
			desire:  nil,
			wantErr: true,
		},
	}

	Convey("FindByAlphaBeta", t, func() {
		for _, c := range cases {
			res, err := testRelationDao.FindByAlphaBeta(ctx, c.alpha, c.beta, false)
			So(err != nil, ShouldEqual, c.wantErr)
			So(res == nil, ShouldEqual, c.desire == nil)
			if c.desire != nil {
				So(res.UserAlpha, ShouldEqual, c.desire.UserAlpha)
				So(res.UserBeta, ShouldEqual, c.desire.UserBeta)
			}
		}
	})
}

func TestRelation_FindUidLinkTo(t *testing.T) {
	testRelations := []*Relation{
		newRelationFromAlphaToBeta(1000, 1002),
		newRelationFromAlphaToBeta(1000, 1003),
		newRelationFromBetaToAlpha(1000, 1004),
		newRelationFromAlphaToBeta(1004, 1002),
		newRelationFromAlphaToBeta(1004, 1005),
		newRelationFromAlphaToBeta(1005, 1002),
		newMutualRelation(1005, 1003),
	}
	for _, c := range testRelations {
		testRelationDao.Insert(ctx, c)
	}

	cases := []struct {
		uid   int64
		wants []int64
	}{
		{
			uid:   1000,
			wants: []int64{1002, 1003},
		},
		{
			uid:   1002,
			wants: []int64{},
		},
		{
			uid:   1004,
			wants: []int64{1000, 1002, 1005},
		},
		{
			uid:   1003,
			wants: []int64{1005},
		},
		{
			uid:   1005,
			wants: []int64{1002, 1003},
		},
		{
			uid:   1001111,
			wants: []int64{},
		},
	}
	Convey("FindUidLinkTo", t, func() {
		for idx, c := range cases {
			res, _, _, err := testRelationDao.FindUidLinkTo(ctx, c.uid, 0, 10)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, len(c.wants))
			slices.Sort(res)
			So(res, ShouldResemble, c.wants)

			// 查两次的结果应该和一次性查出来的结果是一样的
			s1, err := testRelationDao.FindAlphaLinkTo(ctx, c.uid)
			So(err, ShouldBeNil)
			s2, err := testRelationDao.FindBetaLinkTo(ctx, c.uid)
			So(err, ShouldBeNil)
			got := xslice.ConcatUniq(s1, s2)
			slices.Sort(got)
			SoMsg(fmt.Sprintf("[%d]. %v, got:%v, s1:%v, s2:%v want:%v", idx, c, got, s1, s2, c.wants), got, ShouldResemble, c.wants)
		}
	})
}

func TestRelation_FindUidGotLinked(t *testing.T) {
	testRelations := []*Relation{
		newRelationFromAlphaToBeta(1000, 1002),
		newRelationFromAlphaToBeta(1000, 1003),
		newRelationFromBetaToAlpha(1000, 1004),
		newRelationFromAlphaToBeta(1004, 1002),
		newRelationFromAlphaToBeta(1004, 1005),
		newRelationFromAlphaToBeta(1005, 1002),
		newMutualRelation(1005, 1003),
	}
	for _, c := range testRelations {
		testRelationDao.Insert(ctx, c)
	}

	cases := []struct {
		uid   int64   // 关注他
		wants []int64 // 谁关注他
	}{
		{
			uid:   1000,
			wants: []int64{1004},
		},
		{
			uid:   1002,
			wants: []int64{1000, 1004, 1005},
		},
		{
			uid:   1004,
			wants: []int64{},
		},
		{
			uid:   1003,
			wants: []int64{1000, 1005},
		},
		{
			uid:   1005,
			wants: []int64{1003, 1004},
		},
		{
			uid:   1001111,
			wants: []int64{},
		},
	}
	Convey("FindUidGotLinked", t, func() {
		for idx, c := range cases {
			res, _, _, err := testRelationDao.FindUidGotLinked(ctx, c.uid, 0, 10)
			debug := fmt.Sprintf("uid: %d, got: %v, want: %v", c.uid, res, c.wants)
			SoMsg(debug, err, ShouldBeNil)
			SoMsg(debug, res, ShouldHaveLength, len(c.wants))
			slices.Sort(res)
			SoMsg(debug, res, ShouldResemble, c.wants)

			// 查两次的结果应该和一次性查出来的结果是一样的
			s1, err := testRelationDao.FindAlphaGotLinked(ctx, c.uid)
			So(err, ShouldBeNil)
			s2, err := testRelationDao.FindBetaGotLinked(ctx, c.uid)
			So(err, ShouldBeNil)
			got := xslice.ConcatUniq(s1, s2)
			slices.Sort(got)
			SoMsg(fmt.Sprintf("[%d]. %v, got:%v, s1:%v, s2:%v want:%v", idx, c, got, s1, s2, c.wants), got, ShouldResemble, c.wants)
		}
	})
}

func TestRelation_Count(t *testing.T) {
	Convey("CountFans", t, func() {
		cnt, err := testRelationDao.CountUidFans(ctx, 1001)
		So(err, ShouldBeNil)
		t.Log(cnt)
	})

	Convey("CountFollowings", t, func() {
		cnt, err := testRelationDao.CountUidFollowings(ctx, 1001)
		So(err, ShouldBeNil)
		t.Log(cnt)
	})
}

func TestRelation_FindAllUidLinkTo(t *testing.T) {
	Convey("FindAllUidLinkTo", t, func() {
		res, err := testRelationDao.FindAllUidLinkTo(ctx, 1005)
		So(err, ShouldBeNil)
		t.Log(res)
	})
}

func TestRelation_BatchFindUidLinkTo(t *testing.T) {
	Convey("BatchFindUidLinkTo", t, func() {
		testRelations := []*Relation{
			newRelationFromAlphaToBeta(1000, 1002),
			newRelationFromAlphaToBeta(1000, 1003),
			newRelationFromBetaToAlpha(1000, 1004),
			newRelationFromAlphaToBeta(1004, 1002),
			newRelationFromAlphaToBeta(1004, 1005),
			newRelationFromAlphaToBeta(1005, 1002),
			newMutualRelation(1005, 1003),
		}
		for _, c := range testRelations {
			err := testRelationDao.Insert(ctx, c)
			So(err, ShouldBeNil)
		}

		// checking
		gots, err := testRelationDao.BatchFindUidLinkTo(ctx, 1000, []int64{1002, 1003, 1004, 1005})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 2)
		So(gots[0].UserBeta, ShouldEqual, 1002)
		So(gots[1].UserBeta, ShouldEqual, 1003)

		gots, err = testRelationDao.BatchFindUidLinkTo(ctx, 1002, []int64{1000, 1003, 1004, 1005})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 0)

		gots, err = testRelationDao.BatchFindUidLinkTo(ctx, 1005, []int64{1000, 1002, 1003, 1004})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 2)
		So(gots[0].UserAlpha, ShouldEqual, 1002)
		So(gots[1].UserAlpha, ShouldEqual, 1003)
	})
}
