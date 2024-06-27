package utils

import (
	"unsafe"
)

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func Bytes2String(s []byte) string {
	return unsafe.String(unsafe.SliceData(s), len(s))
}
