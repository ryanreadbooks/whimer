package rules

import (
	"testing"
)

func TestShard(t *testing.T) {
	for _, s := range shards {
		t.Log(s)
	}
}

func TestGetKey(t *testing.T) {
	t.Log(PreviewKey("a"))
	t.Log(PreviewKey("A"))
	t.Log(PreviewKey("b"))
	t.Log(PreviewKey("e"))
	t.Log(PreviewKey("c"))
	t.Log(PreviewKey("e"))
	t.Log(PreviewKey("f"))
}