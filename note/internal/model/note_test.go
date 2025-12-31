package model

import "testing"

func TestFormatNoteVideoKey(t *testing.T) {
	key := FormatNoteVideoKey("hello.mp4", SupportedVideoH264Suffix)
	t.Log(key)

	key = FormatNoteVideoKey("hello", SupportedVideoH264Suffix)
	t.Log(key)

	key = FormatNoteVideoKey("hello.mp5", SupportedVideoH265Suffix)
	t.Log(key)
}
