package xerror

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCommon_ShouldLog(t *testing.T) {
	cases := []struct {
		err    error
		expect bool
	}{
		{
			err:    nil,
			expect: false,
		},
		{
			err:    ErrArgs,
			expect: false,
		},
		{
			err:    ErrInternal,
			expect: true,
		},
		{
			err:    ErrInternal.Msg("test internal error"),
			expect: true,
		},
		{
			err:    Wrap(ErrInvalidArgs),
			expect: false,
		},
		{
			err:    Wrap(ErrInternal),
			expect: true,
		},
		{
			err:    Wrapf(Wrap(ErrPermission), "test permission denied"),
			expect: false,
		},
		{
			err:    Wrapf(Wrap(ErrInternal), "test internal error"),
			expect: true,
		},
		{
			err:    status.Error(codes.InvalidArgument, "invalid arg"),
			expect: false,
		},
		{
			err:    status.Error(codes.Internal, "internal"),
			expect: true,
		},
		{
			err:    Wrap(status.Error(codes.InvalidArgument, "invalid arg")),
			expect: false,
		},
		{
			err:    Wrap(status.Error(codes.Internal, "internal err")),
			expect: true,
		},
		{
			err:    Wrapf(Wrap(status.Error(codes.PermissionDenied, "permdenied")), "pg perm"),
			expect: false,
		},
		{
			err:    Wrapf(Wrap(status.Error(codes.Internal, "internal")), "pg internal"),
			expect: true,
		},
	}

	Convey("ShouldLogTest", t, func() {
		for _, c := range cases {
			got := ShouldLogError(c.err)
			SoMsg(fmt.Sprintf("err = %s", c.err), got, ShouldEqual, c.expect)
		}
	})
}
