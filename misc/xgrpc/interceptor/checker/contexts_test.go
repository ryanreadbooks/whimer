package checker

import (
	"testing"
)

func TestForceSkipUidCheck(t *testing.T) {
	ctx := ForceSkipUidCheck(t.Context())

	v := GetForceSkipUidCheck(ctx)
	if !v {
		t.Fail()
	}
}
