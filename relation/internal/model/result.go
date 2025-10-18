package model

import "github.com/ryanreadbooks/whimer/relation/internal/infra/dao"

type ListResult struct {
	NextOffset int64
	HasMore    bool
}

type UidAndTime struct {
	Uid  int64 // 用户id
	Time int64 // 关注时间
}

func NewUidAndTimeFrom(d dao.UidWithTime) UidAndTime {
	return UidAndTime{
		Uid:  d.Uid,
		Time: d.Time,
	}
}

func NewUidAndTimeSliceFrom(ds []dao.UidWithTime) []UidAndTime {
	res := make([]UidAndTime, 0, len(ds))
	for _, d := range ds {
		res = append(res, NewUidAndTimeFrom(d))
	}
	return res
}

func UidsFromUidAndTimeSlice(uts []UidAndTime) []int64 {
	res := make([]int64, 0, len(uts))
	for _, ut := range uts {
		res = append(res, ut.Uid)
	}
	return res
}

func TimesFromUidAndTimeSlice(uts []UidAndTime) []int64 {
	res := make([]int64, 0, len(uts))
	for _, ut := range uts {
		res = append(res, ut.Time)
	}
	return res
}
func UidsSliceTimeSliceFrom(uts []UidAndTime) ([]int64, []int64) {
	uids := make([]int64, len(uts))
	times := make([]int64, len(uts))
	for i, ut := range uts {
		uids[i] = ut.Uid
		times[i] = ut.Time
	}
	return uids, times
}