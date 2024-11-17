package xtime

import (
	"time"

	"gopkg.in/yaml.v3"
)

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

var (
	Hour = time.Hour
	Day  = 24 * Hour
	Week = 7 * Day
)

func WeekJitter(t time.Duration) time.Duration {
	return Week + JitterDuration(t)
}

func WeekJitterSec(t time.Duration) int {
	return int((Week + JitterDuration(t)).Seconds())
}

func HourJitter(t time.Duration) time.Duration {
	return Hour + JitterDuration(t)
}

func HourJitterSec(t time.Duration) int {
	return int((Hour + JitterDuration(t)).Seconds())
}

func DayJitter(t time.Duration) time.Duration {
	return Day + JitterDuration(t)
}

func DayJitterSec(t time.Duration) int {
	return int((Day + JitterDuration(t)).Seconds())
}
