package xerror

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/smartystreets/goconvey/convey"
)

func TestWrap(t *testing.T) {
	err := Propagate(ErrArgs)
	fmt.Printf("%v", err)
}

func TestApiCall(t *testing.T) {
	err := api()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func dao() error {
	if n := rand.Intn(2); n >= 0 {
		return PropagateMsg(ErrArgs, "dao rand error: %d", n)
	}
	return nil
}

func service2() error {
	err := service()
	if err != nil {
		return PropagateMsg(err, "service2 error")
	}

	return nil
}

func service() error {
	err := dao()
	if err != nil {
		return PropagateMsg(err, "servier error, hello world, id:%d", rand.Intn(123))
	}

	return nil
}

func api() error {
	err := service2()
	return err
}

func TestWrap_UnwindFrames(t *testing.T) {
	convey.Convey("UnwindFrames\n", t, func() {
		sts := UnwindFrames(nil)
		convey.So(sts, convey.ShouldBeEmpty)
		sts = UnwindFrames(api())
		convey.So(sts, convey.ShouldNotBeEmpty)
		fmt.Println(stacktrace.FormatFrames(sts))

		fmt.Println(sts.FormatFuncs())
		fmt.Println(sts.FormatLines())
	})
}

func TestWrap_HasFramesHold(t *testing.T) {
	convey.Convey("HasFramesHold\n", t, func() {
		hold := HasFramesHold(nil)
		convey.So(hold, convey.ShouldBeFalse)
		hold = HasFramesHold(api())
		convey.So(hold, convey.ShouldBeTrue)
	})
}

func TestWrap_Log(t *testing.T) {
	convey.Convey("Log ErrProxy\n", t, func() {
		err := PropagateMsg(
			PropagateMsg(
				Propagate(ErrInvalidArgs).WithField("dao", "one").WithField("dao", "two").WithExtra("query", "select *"),
				"level 1",
			).WithExtra("level", 1).WithField("level1", 12),
			"level 2",
		).WithCtx(context.Background()).WithField("level11", "two")

		t.Log(err)
		t.Log(err.Fields())
		t.Log(err.Extra())
	})
}

func TestWrap_UnwindMsg(t *testing.T) {
	cases := []struct {
		err error
	}{
		{
			err: PropagateMsg(
				PropagateMsg(
					Propagate(ErrInvalidArgs).WithField("dao", "one").WithField("dao", "two").WithExtra("query", "select *"),
					"level 1",
				).WithExtra("level", 1).WithField("level1", 12),
				"level 2",
			).WithCtx(context.Background()).WithField("level11", "two"),
		},
		{
			err: nil,
		},
	}
	convey.Convey("Log ErrProxy\n", t, func() {
		for _, c := range cases {
			msg := UnwindMsg(c.err)
			t.Log(msg)
		}
	})
}
