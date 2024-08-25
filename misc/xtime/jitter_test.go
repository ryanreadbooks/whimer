package xtime

import (
	"testing"
	"time"
)

func TestJitter(t *testing.T) {
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))
	t.Log(JitterDuration(time.Hour))

	week := time.Hour * 24 * 7
	t.Log(week)
	t.Log(week + JitterDuration(2 * time.Hour))
	t.Log(week + JitterDuration(2 * time.Hour))
	t.Log(week + JitterDuration(2 * time.Hour))
	t.Log(week + JitterDuration(2 * time.Hour))
	t.Log(week + JitterDuration(2 * time.Hour))
	t.Log(week + JitterDuration(2 * time.Hour))
}
