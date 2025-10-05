package uuid

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

type UUID struct {
	uuid.UUID
}

func (u UUID) Value() (driver.Value, error) {
	return u.UUID[:], nil
}

func (u UUID) Duplicate() UUID {
	dst := [16]byte{}
	copy(dst[:], u.UUID[:])
	return UUID{dst}
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