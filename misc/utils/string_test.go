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

func TestByteCandis(t *testing.T) {
	t.Log(byteCandis)
	t.Log(RandomByte(1))
	t.Log(RandomByte(10))
	t.Log(RandomByte(5))
	t.Log(RandomByte(100))
	t.Log(RandomByte(48))
	t.Log(RandomByte(32))
}

func TestSecureRandomString(t *testing.T) {
	s, _ := SecureRandomString(10)
	t.Log(s, len(s))
}
