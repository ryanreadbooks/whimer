package utils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"strings"
	"unsafe"
)

var (
	candis        = genCandis(false)
	candisLen     = len(candis)
	byteCandis    = genByteCandis()
	byteCandisLen = len(byteCandis)

	passCandis    = genCandis(true)
	passCandisLen = len(passCandis)
)

func genCandis(pass bool) string {
	var bd strings.Builder
	bd.Grow(256)
	for i := 'a'; i <= 'z'; i++ {
		bd.WriteRune(i)
	}

	for i := 'A'; i <= 'Z'; i++ {
		bd.WriteRune(i)
	}

	for i := '0'; i <= '9'; i++ {
		bd.WriteRune(i)
	}

	if !pass {
		bd.WriteString("+-*/=~!@#$%^&_<>?:'[]{}|.")
	} else {
		bd.WriteString("!@#$%^&*()")
	}

	return bd.String()
}

func genByteCandis() []byte {
	return []byte(genCandis(false))
}

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func Bytes2String(s []byte) string {
	return unsafe.String(unsafe.SliceData(s), len(s))
}

// 生成长度为size的随机字符串
func RandomString(size int) string {
	var bd strings.Builder
	bd.Grow(size)

	for range size {
		bd.WriteByte(candis[rand.Intn(candisLen)])
	}

	return bd.String()
}

func RandomByte(size int) []byte {
	var bd bytes.Buffer
	bd.Grow(size)
	for range size {
		bd.WriteByte(byteCandis[rand.Intn(byteCandisLen)])
	}

	return bd.Bytes()
}

func CryptoRandomString(size int) (s string, err error) {
	var buf = make([]byte, size)
	_, err = crand.Read(buf)
	if err != nil {
		return
	}

	s = hex.EncodeToString(buf)
	return
}

// 生成长度为size的随机密码
func RandomPass(size int) string {
	var bd strings.Builder
	bd.Grow(size)

	for i := 0; i < size; i++ {
		bd.WriteByte(passCandis[rand.Intn(passCandisLen)])
	}

	return bd.String()
}
