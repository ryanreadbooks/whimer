package xsql

import (
	"testing"
)

func TestSQLState(t *testing.T) {
	s := [5]uint8{50, 50, 48, 48, 51}
	t.Log(SQLStateEqual(s, SQLStateOutOfRange))
}