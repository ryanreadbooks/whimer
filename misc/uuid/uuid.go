package uuid

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

var (
	zeroUUID = UUID{uuid.Nil}
	maxUUID  = UUID{uuid.Max}
)

type UUID struct {
	uuid.UUID
}

func EmptyUUID() UUID {
	return zeroUUID
}

func MaxUUID() UUID {
	return maxUUID
}

func (u UUID) Value() (driver.Value, error) {
	return u.UUID[:], nil
}

func (u UUID) Duplicate() UUID {
	dst := [16]byte{}
	copy(dst[:], u.UUID[:])
	return UUID{dst}
}

func (u UUID) Time() time.Time {
	t := u.UUID.Time()
	sec, nesc := t.UnixTime() // unix time with second and nanosec
	return time.Unix(sec, nesc)
}

func (u UUID) UnixSec() int64 {
	return u.Time().Unix()
}

func (u UUID) UnixMill() int64 {
	return u.Time().UnixMilli()
}

func ParseString(s string) (UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return EmptyUUID(), err
	}
	return UUID{u}, nil
}

func NewUUID() UUID {
	return UUID{uuid.Must(uuid.NewV7())}
}

// compare u to o, return -1 if u < o, 0 if u == o, 1 if u > o
func (u UUID) Compare(o UUID) int {
	for idx := range 16 {
		if u.UUID[idx] < o.UUID[idx] {
			return -1
		} else if u.UUID[idx] > o.UUID[idx] {
			return 1
		}
	}
	return 0
}

func (u UUID) GreaterThan(o UUID) bool {
	return u.Compare(o) > 0
}

func (u UUID) NotEqualsTo(o UUID) bool {
	return u.Compare(o) != 0
}

func (u UUID) EqualsTo(o UUID) bool {
	return u.Compare(o) == 0
}

func (u UUID) LessThan(o UUID) bool {
	return u.Compare(o) < 0
}

func (u UUID) IsZero() bool {
	return u.EqualsTo(zeroUUID)
}

func (u UUID) IsMax() bool {
	return u.EqualsTo(maxUUID)
}
