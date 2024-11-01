package xsql

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNoRecord   = xerror.ErrNotFound.Msg("no record found")
	ErrDuplicate  = xerror.ErrInternal.Msg("duplicate entry")
	ErrOutOfRange = xerror.ErrInternal.Msg("out of range")
)

// 转换not found和duplicate entry两种错误
func ConvertError(err error) error {
	switch err {
	case nil:
		return nil
	case sqlx.ErrNotFound:
		return ErrNoRecord
	default:
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			if mysqlErr.Number == 1062 {
				return ErrDuplicate
			} else if SQLStateEqual(mysqlErr.SQLState, SQLStateOutOfRange) {
				return ErrOutOfRange
			}
		}

		// 其它db错误全部视为5xx
		return xerror.Wrapf(xerror.ErrInternal, err.Error())
	}
}

func IsMildErr(err error) bool {
	if err == nil {
		return true
	}
	return errors.Is(err, ErrNoRecord) || errors.Is(err, ErrDuplicate)
}

func IsCriticalErr(err error) bool {
	if err == nil {
		return false
	}

	return !IsMildErr(err)
}

func IsDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNoRecord)
}
