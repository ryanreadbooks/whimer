package xtime

import (
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type SDuration string

func (s SDuration) Duration() time.Duration {
	d, _ := time.ParseDuration(string(s))
	return d
}

// 可以解析1m 2h等的时间格式
type Duration time.Duration

// 实现yaml.Unmarshaler接口
func (t *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	r, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*t = Duration(r)

	return nil
}

func (t *Duration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	r, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*t = Duration(r)

	return nil
}

const (
	Sec       = 1
	MinuteSec = 60 * Sec
	HourSec   = 60 * MinuteSec
	DaySec    = 24 * HourSec
	WeekSec   = 7 * DaySec
)

var (
	Hour = time.Hour
	Day  = 24 * Hour
	Week = 7 * Day
)

func WeekJitter(t time.Duration) time.Duration {
	return Week + JitterDuration(t)
}

func NWeekJitter(n int, t time.Duration) time.Duration {
	return time.Duration(n)*Week + JitterDuration(t)
}

func WeekJitterSec(t time.Duration) int {
	return int((Week + JitterDuration(t)).Seconds())
}

func HourJitter(t time.Duration) time.Duration {
	return Hour + JitterDuration(t)
}

func NHourJitter(n int, t time.Duration) time.Duration {
	return time.Duration(n)*Hour + JitterDuration(t)
}

func HourJitterSec(t time.Duration) int {
	return int((Hour + JitterDuration(t)).Seconds())
}

func DayJitter(t time.Duration) time.Duration {
	return Day + JitterDuration(t)
}

func NDayJitter(n int, t time.Duration) time.Duration {
	return time.Duration(n)*Day + JitterDuration(t)
}

func DayJitterSec(t time.Duration) int {
	return int((Day + JitterDuration(t)).Seconds())
}
