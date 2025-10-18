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
	err := Wrap(ErrArgs)
	fmt.Printf("%v", err)
}

func TestApiCall(t *testing.T) {
	err := testApi()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func testConverErr(err error) error {
	// 其它db错误全部视为5xx
	return Wrapf(ErrInternal, "%s", err.Error())
}

func dao() error {
	if n := rand.Intn(2); n >= 0 {
		// return Wrapf(ErrArgs, "dao rand error: %d", n)
		return Wrap(testConverErr(fmt.Errorf("unknown db error")))
	}
	return nil
}

func service2() error {
	err := service()
	if err != nil {
		return Wrap(err)
	}

	return nil
}

func service() error {
	err := dao()
	if err != nil {
		return Wrapf(err, "hello world id:%d", rand.Intn(123))
	}

	return nil
}

func testApi() error {
	err := service2()
	return err
}

func TestWrap_UnwindFrames(t *testing.T) {
	convey.Convey("UnwindFrames\n", t, func() {
		sts := UnwrapFrames(nil)
		convey.So(sts, convey.ShouldBeEmpty)
		sts = UnwrapFrames(testApi())
		convey.So(sts, convey.ShouldNotBeEmpty)
		fmt.Println(stacktrace.FormatFrames(sts))

		fmt.Println(sts.FormatFuncs())
		fmt.Println(sts.FormatLines())
	})
}

func TestWrap_HasFramesHold(t *testing.T) {
	convey.Convey("HasFramesHold\n", t, func() {
		hold := FramesWrapped(nil)
		convey.So(hold, convey.ShouldBeFalse)
		hold = FramesWrapped(testApi())
		convey.So(hold, convey.ShouldBeTrue)
	})
}

func TestWrap_Log(t *testing.T) {
	convey.Convey("Log ErrProxy\n", t, func() {
		err := Wrapf(
			Wrapf(
				Wrap(ErrInvalidArgs).WithField("dao", "one").WithField("dao", "two").WithExtra("query", "select *"),
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
			err: Wrapf(
				Wrapf(
					Wrap(ErrInvalidArgs).WithField("dao", "one").WithField("dao", "two").WithExtra("query", "select *"),
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
			msg, underErr := UnwrapMsg(c.err)
			t.Log(msg)
			t.Log(underErr)
		}
	})
}

func TestStripFrames(t *testing.T) {
	err := testApi()
	sts := UnwrapFrames(err)
	fmt.Println(stacktrace.FormatFrames(sts))
	t.Log(err)

	t.Log("-------------------------")

	err2 := StripFrames(err)
	sts2 := UnwrapFrames(err2)
	fmt.Println(stacktrace.FormatFrames(sts2))
	t.Log(err2)

	err3 := StripFrames(nil)
	t.Log(err3)

	err3 = StripFrames(fmt.Errorf("hello"))
	t.Log(err3)
}
