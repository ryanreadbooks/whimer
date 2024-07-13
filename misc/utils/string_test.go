package utils

import (
	"testing"
)

func TestCandis(t *testing.T) {
	t.Log(candis)
	t.Log(RandomString(1))
	t.Log(RandomString(10))
	t.Log(RandomString(5))
	t.Log(RandomString(100))
	t.Log(RandomString(48))
	t.Log(RandomString(16))
}
